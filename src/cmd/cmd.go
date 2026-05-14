package cmd

import "github.com/zeropsio/zcli/src/kongCmd"

// ExecuteCmd is the binary entry point. Routes through the kong dispatcher
// against the kongRootCLI grammar. The cobra-era root + cmdBuilder linger
// for `project scope` only — that one gets retired in a future refactor.
func ExecuteCmd() {
	kongCmd.ExecuteRootCmd(&kongRootCLI{})
}
