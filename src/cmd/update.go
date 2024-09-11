package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
    "strings"
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

func updateCmd() error {
    checkForUpdates := func() error {
        currentVersion, err := getCurrentVersion()
        if err != nil {
            return fmt.Errorf("failed to get current version: %w", err)
        }

        latestVersion, err := getLatestVersion()
        if err != nil {
            return err
        }

        if strings.Compare(currentVersion, latestVersion) < 0 {
            return promptForUpdate(latestVersion)
        }
        return nil
    }

    getLatestVersion := func() (string, error) {
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

    promptForUpdate := func(latestVersion string) error {
        fmt.Printf("A new version (v%s) is available. Do you want to update now? (y/n): ", latestVersion)
        var response string
        _, err := fmt.Scanln(&response)
        if err != nil {
            return err
        }

        if strings.ToLower(response) == "y" {
            return updateCLI()
        }

        fmt.Println("Update skipped.")
        return nil
    }

    updateCLI := func() error {
        fmt.Println("Updating to the latest version...")
        cmd := exec.Command("npm", "install", "-g", npmPackageName)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        err := cmd.Run()
        if err != nil {
            return err
        }
        fmt.Println("Update successful.")
        return nil
    }

    return checkForUpdates()
}

func main() {
    if err := updateCmd(); err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
}
