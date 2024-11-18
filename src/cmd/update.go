package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

// other necessary imports

func updateCmd() *cmdBuilder.Cmd {
	// check existing zcli version
	return cmdBuilder.NewCmd().
		Use("update").
		Short(i18n.T(i18n.CmdDescUpdate)).
		HelpFlag(i18n.T(i18n.CmdHelpUpdate)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			// print the current version of zcli
			if version != "" {
				latestVersion, err := getLatestGitHubRelease(ctx)
				if err != nil {
					return err
				}

				if latestVersion.TagName != version {
					fmt.Println("There is a new version available:", latestVersion.TagName)
					fmt.Println("Do you want to update? (y/n)")
					var input string
					if _, err := fmt.Scanln(&input); err != nil {
						fmt.Println("Failed to read input:", err)
						return err
					}

					if input == "y" {
						target := determineTargetArchitecture()
						if err := downloadAndInstallZCLI(ctx, target); err != nil {
							return err
						}
						fmt.Println("zCLI was updated successfully to", latestVersion.TagName)
					} else {
						fmt.Println("Update canceled.")
					}
				} else {
					fmt.Println("You are using the latest version of zcli")
				}
			} else {
				fmt.Println("You are using the development environment of zcli")
			}
			return nil
		})
}

type GitHubRelease struct {
	TagName string `json:"tagName"`
	Body    string `json:"body"`
}

func getLatestGitHubRelease(ctx context.Context) (GitHubRelease, error) {
	// GitHub repository details
	repoOwner := "zeropsio"
	repoName := "zcli"

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	client := http.Client{
		Timeout: 4 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return GitHubRelease{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return GitHubRelease{}, nil
		}
		return GitHubRelease{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("erorr")
		}
	}()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return GitHubRelease{}, err
	}
	return release, nil
}

func determineTargetArchitecture() string {
	switch runtime.GOOS + " " + runtime.GOARCH {
	case "darwin amd64":
		return "darwin-amd64"
	case "darwin arm64":
		return "darwin-arm64"
	case "linux 386":
		return "linux-i386"
	default:
		return "linux-amd64"
	}
}

func downloadAndInstallZCLI(ctx context.Context, target string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	binDir := fmt.Sprintf("%s/.local/bin", homeDir)
	binPath := fmt.Sprintf("%s/zcli", binDir)

	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", binDir, err)
		}
	}

	zcliURI := fmt.Sprintf("https://github.com/zeropsio/zcli/releases/latest/download/zcli-%s", target)
	curlCmd := fmt.Sprintf("curl --fail --location --progress-bar --output %s %s", binPath, zcliURI)
	cmd := exec.CommandContext(ctx, "sh", "-c", curlCmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download zcli: %w", err)
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("failed to make zcli executable: %w", err)
	}

	fmt.Printf("zCLI was installed successfully to %s\n", binPath)
	return nil
}
