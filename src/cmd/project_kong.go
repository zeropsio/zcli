package cmd

// ProjectCmd groups project subcommands. Commands migrate into here as
// phase 3 progresses.
type ProjectCmd struct {
	Create        ProjectCreateCmd        `cmd:"" help:"Creates a new project for specified org."`
	Delete        ProjectDeleteCmd        `cmd:"" help:"Deletes the selected project."`
	List          ProjectListCmd          `cmd:"" help:"List of all imported projects."`
	ProjectImport ProjectImportCmd        `cmd:"project-import" help:"Import project from a YAML spec."`
	ServiceImport ProjectServiceImportCmd `cmd:"service-import" help:"Import service(s) into the scoped project from a YAML spec."`
}
