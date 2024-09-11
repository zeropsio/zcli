package main

import (
	"github.com/zeropsio/zcli/src/cmd"
	"/src/cmd/update.go"
)

func main() {

    go func() {
        if err := updateCmd(); err != nil {
            fmt.Println("Error checking for updates:", err)
        }
    }()

	cmd.ExecuteCmd()
}
