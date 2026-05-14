package cmd

// ServiceCmd groups service subcommands. Commands migrate into here as
// phase 3 progresses.
type ServiceCmd struct {
	List ServiceListCmd `cmd:"" help:"List of all services in the scoped project."`
}
