package cmd

// ServiceCmd groups service subcommands. Commands migrate into here as
// phase 3 progresses.
type ServiceCmd struct {
	Create          ServiceCreateCmd          `cmd:"" help:"Create a new service for the scoped project."`
	Delete          ServiceDeleteCmd          `cmd:"" help:"Delete the scoped service."`
	Deploy          ServiceDeployCmd          `cmd:"" help:"Deploy a new app version from local files."`
	EnableSubdomain ServiceEnableSubdomainCmd `cmd:"enable-subdomain" help:"Enable zerops subdomain for the scoped service."`
	List            ServiceListCmd            `cmd:"" help:"List of all services in the scoped project."`
	Log             ServiceLogCmd             `cmd:"" help:"Stream logs from the scoped service."`
	Start           ServiceStartCmd           `cmd:"" help:"Start the scoped service."`
	Stop            ServiceStopCmd            `cmd:"" help:"Stop the scoped service."`
}
