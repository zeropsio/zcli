package protocol

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	validTokenPattern  = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	validTargetPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

type DefaultHandler struct{}

func NewDefaultHandler() Handler {
	return &DefaultHandler{}
}

func (h *DefaultHandler) Handle(ctx context.Context, protocolURL *url.URL) error {
	if protocolURL.Scheme != "zcli" {
		return fmt.Errorf("unsupported protocol scheme: %s", protocolURL.Scheme)
	}

	command := strings.TrimPrefix(protocolURL.Path, "/")
	if command == "" {
		command = protocolURL.Host
	}

	switch command {
	case "login":
		return h.handleLogin(ctx, protocolURL)
	case "open":
		return h.handleOpen(ctx, protocolURL)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (h *DefaultHandler) handleLogin(ctx context.Context, u *url.URL) error {
	fmt.Println("Handling login request")

	query := u.Query()
	token := query.Get("token")

	if err := h.validateToken(token); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	fmt.Println("Login token received")
	return nil
}

func (h *DefaultHandler) handleOpen(ctx context.Context, u *url.URL) error {
	fmt.Println("Handling open request")

	query := u.Query()
	target := query.Get("target")

	if err := h.validateTarget(target); err != nil {
		return fmt.Errorf("invalid target: %w", err)
	}

	fmt.Printf("Opening target: %s\n", target)
	return nil
}

func (h *DefaultHandler) validateToken(token string) error {
	if len(token) == 0 {
		return fmt.Errorf("token is required")
	}

	if len(token) < 32 || len(token) > 512 {
		return fmt.Errorf("token length invalid")
	}

	if !validTokenPattern.MatchString(token) {
		return fmt.Errorf("token contains invalid characters")
	}

	return nil
}

func (h *DefaultHandler) validateTarget(target string) error {
	if len(target) == 0 {
		return fmt.Errorf("target cannot be empty")
	}

	if len(target) > 256 {
		return fmt.Errorf("target too long")
	}

	if !validTargetPattern.MatchString(target) {
		return fmt.Errorf("target contains invalid characters")
	}

	return nil
}
