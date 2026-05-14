package cmd

// ServiceCmd groups service subcommands. Commands migrate into here as
// phase 3 progresses.
type ServiceCmd struct {
	EnableSubdomain ServiceEnableSubdomainCmd `cmd:"enable-subdomain" help:"Enable zerops subdomain for the scoped service."`
	List            ServiceListCmd            `cmd:"" help:"List of all services in the scoped project."`
	Start           ServiceStartCmd           `cmd:"" help:"Start the scoped service."`
	Stop            ServiceStopCmd            `cmd:"" help:"Stop the scoped service."`
}
