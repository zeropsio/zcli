package cmd

// ProjectCmd groups project subcommands. Commands migrate into here as
// phase 3 progresses; first occupant is List.
type ProjectCmd struct {
	List ProjectListCmd `cmd:"" help:"List of all imported projects."`
}
