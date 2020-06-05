package main

import (
	"github.com/zerops-io/zcli/cmd"
)

var (
	Token string
)

func main() {
	cmd.ExecuteRootCmd(Token)
}
