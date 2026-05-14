package cmd

// kongRootCLI is the kong grammar root. Commands migrate from rootCmd()
// (cobra) into here one at a time during phase 3. When every command lives
// here, ExecuteCmd flips to kong and rootCmd() goes away.
type kongRootCLI struct {
	Version VersionCmd `cmd:"" help:"Print zcli version."`
}
