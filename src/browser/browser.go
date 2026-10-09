package browser

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

var ErrNoBrowser = errors.New("no browser available")

// Open starts the default browser with rawURL and does not wait for it.
// It returns an error when no browser can be started.
func Open(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errors.Wrap(err, "parse url")
	}
	// rundll32 and open also start local files and programs, so only web links are passed on.
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.Errorf("unsupported url scheme %q", parsed.Scheme)
	}

	h := host{
		goos:     runtime.GOOS,
		getenv:   os.Getenv,
		readFile: os.ReadFile,
		lookPath: exec.LookPath,
	}
	argv, err := h.command(rawURL)
	if err != nil {
		return err
	}

	// The process is not tied to a caller context, so cancelling the login never kills a browser that is starting up.
	// Nil stdin, stdout and stderr connect to the null device, so launcher warnings never reach the terminal.
	cmd := exec.CommandContext(context.Background(), argv[0], argv[1:]...)
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "start %s", argv[0])
	}
	// The launcher exit code says nothing about the browser, so the process is only reaped.
	go func() { _ = cmd.Wait() }()
	return nil
}

type host struct {
	goos     string
	getenv   func(string) string
	readFile func(string) ([]byte, error)
	lookPath func(string) (string, error)
}

func (h host) command(rawURL string) ([]string, error) {
	switch h.goos {
	case "darwin":
		return []string{"open", rawURL}, nil
	case "windows":
		// rundll32 hands the url to ShellExecute without a shell, so every & stays intact.
		return []string{"rundll32", "url.dll,FileProtocolHandler", rawURL}, nil
	case "linux":
		// WSLg sets DISPLAY, so xdg-open would start a Linux browser inside WSL instead of the Windows one.
		if h.isWSL() {
			return h.wslCommand(rawURL)
		}
		return h.xdgCommand(rawURL)
	case "freebsd", "openbsd", "netbsd", "dragonfly", "solaris", "illumos":
		return h.xdgCommand(rawURL)
	default:
		return nil, errors.Wrapf(ErrNoBrowser, "unsupported OS %s", h.goos)
	}
}

func (h host) isWSL() bool {
	if h.getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	osRelease, err := h.readFile("/proc/sys/kernel/osrelease")
	return err == nil && strings.Contains(strings.ToLower(string(osRelease)), "microsoft")
}

// wslCommand never uses explorer.exe, which opens the Documents folder for some URLs.
// The absolute rundll32.exe path covers `appendWindowsPath = false` in /etc/wsl.conf, and a Linux browser in WSLg is the last resort.
func (h host) wslCommand(rawURL string) ([]string, error) {
	if _, err := h.lookPath("wslview"); err == nil {
		return []string{"wslview", rawURL}, nil
	}
	for _, name := range []string{"rundll32.exe", "/mnt/c/Windows/System32/rundll32.exe"} {
		if _, err := h.lookPath(name); err == nil {
			return []string{name, "url.dll,FileProtocolHandler", rawURL}, nil
		}
	}
	return h.xdgCommand(rawURL)
}

// xdgCommand never falls back to x-www-browser, www-browser or $BROWSER, which may start a text browser that takes over the terminal.
func (h host) xdgCommand(rawURL string) ([]string, error) {
	if h.getenv("DISPLAY") == "" && h.getenv("WAYLAND_DISPLAY") == "" {
		return nil, errors.Wrap(ErrNoBrowser, "no graphical session, DISPLAY and WAYLAND_DISPLAY are unset")
	}
	if _, err := h.lookPath("xdg-open"); err != nil {
		return nil, errors.Wrap(ErrNoBrowser, "xdg-open not found")
	}
	return []string{"xdg-open", rawURL}, nil
}
