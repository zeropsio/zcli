package protocol

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Registration manages the registration of custom protocol handlers with the operating system
type Registration interface {
	Register(ctx context.Context) error
	Unregister(ctx context.Context) error
	IsRegistered(ctx context.Context) (bool, error)
}

// Handler processes custom protocol URLs
type Handler interface {
	Handle(ctx context.Context, protocolURL *url.URL) error
}

func NewRegistration() Registration {
	return &registration{}
}

func GetExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	absPath, err := filepath.Abs(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}

func OpenURL(urlStr string) error {
	if err := validateURL(urlStr); err != nil {
		return err
	}
	
	var cmd *exec.Cmd
	
	switch {
	case isCommand("xdg-open"):
		cmd = exec.Command("xdg-open", urlStr)
	case isCommand("open"):
		cmd = exec.Command("open", urlStr)
	case isCommand("start"):
		cmd = exec.Command("cmd", "/c", "start", urlStr)
	}
	
	if cmd != nil {
		return cmd.Start()
	}
	
	return nil
}

var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
	"file":  true,
	"zcli":  true,
}

func validateURL(urlStr string) error {
	if len(urlStr) == 0 {
		return fmt.Errorf("URL cannot be empty")
	}
	
	if len(urlStr) > 2048 {
		return fmt.Errorf("URL too long")
	}
	
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	
	if !allowedSchemes[parsedURL.Scheme] {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}
	
	if strings.Contains(urlStr, "../") || strings.Contains(urlStr, "..\\") {
		return fmt.Errorf("path traversal detected")
	}
	
	return nil
}

func isCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}