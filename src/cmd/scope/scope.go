package scope

var Project *project
var Service *service

func init() {
	Project = &project{}
	Service = &service{
		parent: Project,
	}
}
