//go:build linux

package protocol

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type registration struct{}

const desktopTemplate = `[Desktop Entry]
Version=1.0
Type=Application
Name=Zerops CLI Protocol Handler
Comment=Handle zcli:// protocol URLs
Exec=%s protocol handle %%u
Icon=terminal
StartupNotify=true
NoDisplay=true
MimeType=x-scheme-handler/zcli;
`

func (r *registration) Register(ctx context.Context) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	appsDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}

	desktopPath := filepath.Join(appsDir, "zcli-protocol-handler.desktop")
	desktopContent := fmt.Sprintf(desktopTemplate, execPath)

	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	if err := r.updateMimeDatabase(ctx); err != nil {
		return fmt.Errorf("failed to update MIME database: %w", err)
	}

	if err := r.updateDesktopDatabase(ctx); err != nil {
		return fmt.Errorf("failed to update desktop database: %w", err)
	}

	return nil
}

func (r *registration) Unregister(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	desktopPath := filepath.Join(homeDir, ".local", "share", "applications", "zcli-protocol-handler.desktop")

	if _, err := os.Stat(desktopPath); err == nil {
		if err := os.Remove(desktopPath); err != nil {
			return fmt.Errorf("failed to remove desktop file: %w", err)
		}
	}

	if err := r.updateMimeDatabase(ctx); err != nil {
		return fmt.Errorf("failed to update MIME database: %w", err)
	}

	if err := r.updateDesktopDatabase(ctx); err != nil {
		return fmt.Errorf("failed to update desktop database: %w", err)
	}

	return nil
}

func (r *registration) IsRegistered(ctx context.Context) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get home directory: %w", err)
	}

	desktopPath := filepath.Join(homeDir, ".local", "share", "applications", "zcli-protocol-handler.desktop")
	_, err = os.Stat(desktopPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check desktop file: %w", err)
	}

	return true, nil
}

func (r *registration) updateMimeDatabase(ctx context.Context) error {
	if !isCommand("update-mime-database") {
		return nil // Silently skip if command not available
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	mimeDir := filepath.Join(homeDir, ".local", "share", "mime")
	if err := os.MkdirAll(mimeDir, 0755); err != nil {
		return fmt.Errorf("failed to create mime directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "update-mime-database", mimeDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update MIME database: %w (output: %s)", err, string(output))
	}

	return nil
}

func (r *registration) updateDesktopDatabase(ctx context.Context) error {
	if !isCommand("update-desktop-database") {
		return nil // Silently skip if command not available
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	appsDir := filepath.Join(homeDir, ".local", "share", "applications")
	cmd := exec.CommandContext(ctx, "update-desktop-database", appsDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update desktop database: %w (output: %s)", err, string(output))
	}

	return nil
}