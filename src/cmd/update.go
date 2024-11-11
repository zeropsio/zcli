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
			latestVersion, err := getLatestGitHubRelease(ctx)
			if err != nil {
				return err
			}

			if latestVersion.TagName != version {
				fmt.Println("There is a new version available:", latestVersion.TagName)
				fmt.Println("Do you want to update? (y/n)")
				var input string
				fmt.Scanln(&input)

				if input == "y" {
					fmt.Println("Updating zcli...")

					// Set the target based on system architecture
					var target string
					switch runtime.GOOS + " " + runtime.GOARCH {
					case "darwin amd64":
						target = "darwin-amd64"
					case "darwin arm64":
						target = "darwin-arm64"
					case "linux 386":
						target = "linux-i386"
					default:
						target = "linux-amd64"
					}

					// Determine the URI for the download based on the target
					var zcliURI = fmt.Sprintf("https://github.com/zeropsio/zcli/releases/latest/download/zcli-%s", target)

					// Define installation path
					binDir := os.ExpandEnv("$HOME/.local/bin")
					binPath := fmt.Sprintf("%s/zcli", binDir)

					// Create binDir if it doesn't exist
					if _, err := os.Stat(binDir); os.IsNotExist(err) {
						if err := os.MkdirAll(binDir, 0755); err != nil {
							return fmt.Errorf("failed to create directory %s: %v", binDir, err)
						}
					}

					// Download zcli binary
					curlCmd := fmt.Sprintf("curl --fail --location --progress-bar --output %s %s", binPath, zcliURI)
					cmd := exec.Command("sh", "-c", curlCmd)

					if err := cmd.Run(); err != nil {
						return fmt.Errorf("failed to download zcli: %v", err)
					}

					// Make binary executable
					if err := os.Chmod(binPath, 0755); err != nil {
						return fmt.Errorf("failed to make zcli executable: %v", err)
					}

					fmt.Println("zCLI was installed successfully to", binPath)
				} else {
					fmt.Println("Update canceled.")
				}
			} else {
				fmt.Println("You are using the latest version of zcli")
			}
			return nil
		})
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
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

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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
