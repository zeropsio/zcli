package main

import (
	"os"

	"github.com/zerops-io/zcli/src/cmd"
)

var (
	Token string
)

func main() {
	cmd.BuiltinToken = Token
	err := cmd.ExecuteCmd()
	if err != nil {
		os.Exit(1)
	}
}
