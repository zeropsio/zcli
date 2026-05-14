package cmd

// kongRootCLI is the kong grammar root. Commands migrate from rootCmd()
// (cobra) into here one at a time during phase 3. When every command lives
// here, ExecuteCmd flips to kong and rootCmd() goes away.
type kongRootCLI struct {
	Login          LoginCmd         `cmd:"" help:"Login into Zerops with generated Zerops token"`
	Logout         LogoutCmd        `cmd:"" help:"Disconnect from VPN and log out from your Zerops account"`
	ShowDebugLogs  ShowDebugLogsCmd `cmd:"show-debug-logs" help:"Display zcli debug logs."`
	Support        SupportCmd       `cmd:"" help:"Display Zerops contact info."`
	Version        VersionCmd       `cmd:"" help:"Print zcli version."`
}
