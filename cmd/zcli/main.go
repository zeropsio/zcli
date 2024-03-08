package main

import (
	"os"

	"github.com/zeropsio/zcli/src/cmd"
)

func main() {
	if cmd.ExecuteCmd() != nil {
		os.Exit(1)
	}
}
