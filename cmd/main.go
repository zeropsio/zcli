package main

import (
	"os"

	"github.com/zeropsio/zcli/src/cmd"
)

func main() {
	err := cmd.ExecuteCmd()
	if err != nil {
		os.Exit(1)
	}
}
