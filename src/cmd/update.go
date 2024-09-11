package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"bufio"
	"strings"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

const npmPackageName = "zcli"

func getCurrentVersion() (string, error) {
	cmd := exec.Command("zcli", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "zcli version") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}
	return "", fmt.Errorf("could not parse zcli version")
}

func getLatestVersion() (string, error) {
	url := "https://registry.npmjs.org/" + npmPackageName
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	latestVersion := data["dist-tags"].(map[string]interface{})["latest"].(string)
	return latestVersion, nil
}

func updateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("update").
		Short(i18n.T(i18n.CmdDescUpdate)).
		HelpFlag(i18n.T(i18n.CmdHelpUpdate)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			currentVersion, err := getCurrentVersion()
			if err != nil {
				return fmt.Errorf("failed to get current version: %w", err)
			}

			latestVersion, err := getLatestVersion()
			if err != nil {
				return err
			}

			if strings.Compare(currentVersion, latestVersion) < 0 {
				cmdData.Stdout.Printf("A new version (v%s) is available. Do you want to update now? (y/n): ", latestVersion)
				var response string
				_, err := fmt.Scanln(&response)
				if err != nil {
					return err
				}

				if strings.ToLower(response) == "y" {
					cmd := exec.Command("npm", "install", "-g", npmPackageName)
					cmd.Stdout = cmdData.Stdout
					cmd.Stderr = cmdData.Stderr
					err := cmd.Run()
					if err != nil {
						return err
					}
					cmdData.Stdout.Println("Update successful. Please restart the CLI.")
				} else {
					cmdData.Stdout.Println("Update skipped.")
				}
			} else {
				cmdData.Stdout.Println("You are already using the latest version.")
			}

			return nil
		})
}
