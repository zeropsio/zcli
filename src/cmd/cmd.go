package cmd

import "github.com/zeropsio/zcli/src/kongCmd"

// ExecuteCmd is the binary entry point. Routes through the kong dispatcher
// against the kongRootCLI grammar.
func ExecuteCmd() {
	kongCmd.ExecuteRootCmd(&kongRootCLI{})
}
