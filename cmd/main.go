package main

import (
	"os"

	"github.com/zeropsio/zcli/src/metaError"

	"github.com/zeropsio/zcli/src/cmd"
)

var (
	Token string
)

func main() {
	cmd.BuiltinToken = Token
	err := cmd.ExecuteCmd()
	if err != nil {
		metaError.Print(err)
		os.Exit(1)
	}
}
