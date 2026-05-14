package cmd

// ProjectCmd groups project subcommands. Commands migrate into here as
// phase 3 progresses.
type ProjectCmd struct {
	Create ProjectCreateCmd `cmd:"" help:"Creates a new project for specified org."`
	List   ProjectListCmd   `cmd:"" help:"List of all imported projects."`
}
