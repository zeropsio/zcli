//go:build darwin

package protocol

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type registration struct{}

const (
	plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>io.zerops.zcli.protocol</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>protocol</string>
		<string>handle</string>
		<string>%%</string>
	</array>
	<key>CFBundleURLTypes</key>
	<array>
		<dict>
			<key>CFBundleURLName</key>
			<string>Zerops CLI Protocol</string>
			<key>CFBundleURLSchemes</key>
			<array>
				<string>zcli</string>
			</array>
		</dict>
	</array>
</dict>
</plist>`
)

func (r *registration) Register(ctx context.Context) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistPath := filepath.Join(launchAgentsDir, "io.zerops.zcli.protocol.plist")
	plistContent := fmt.Sprintf(plistTemplate, execPath)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	cmd := exec.CommandContext(ctx, "launchctl", "load", plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to load plist: %w (output: %s)", err, string(output))
	}

	_ = r.registerLSHandlers(ctx)

	return nil
}

func (r *registration) Unregister(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "io.zerops.zcli.protocol.plist")

	if _, err := os.Stat(plistPath); err == nil {
		cmd := exec.CommandContext(ctx, "launchctl", "unload", plistPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to unload plist: %w (output: %s)", err, string(output))
		}

		if err := os.Remove(plistPath); err != nil {
			return fmt.Errorf("failed to remove plist file: %w", err)
		}
	}

	_ = r.unregisterLSHandlers(ctx)

	return nil
}

func (r *registration) IsRegistered(ctx context.Context) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get home directory: %w", err)
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "io.zerops.zcli.protocol.plist")
	_, err = os.Stat(plistPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check plist file: %w", err)
	}

	cmd := exec.CommandContext(ctx, "launchctl", "list", "io.zerops.zcli.protocol")
	err = cmd.Run()
	return err == nil, nil
}

func (r *registration) registerLSHandlers(ctx context.Context) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.CommandContext(ctx, "/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister", "-f", execPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "No such file or directory") {
			return fmt.Errorf("failed to register with Launch Services: %w (output: %s)", err, string(output))
		}
	}

	return nil
}

func (r *registration) unregisterLSHandlers(ctx context.Context) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.CommandContext(ctx, "/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister", "-u", execPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "No such file or directory") {
			return fmt.Errorf("failed to unregister with Launch Services: %w (output: %s)", err, string(output))
		}
	}

	return nil
}
