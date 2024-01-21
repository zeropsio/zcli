package cmdBuilder

import (
	"context"
	"slices"
)

// FIXME - janhajek do we need interface?
type Dependency interface {
	AddCommandFlags(*Cmd)
	LoadSelectedScope(ctx context.Context, cmd *Cmd, cmdData *LoggedUserCmdData) error
	GetParent() Dependency
}

type commonDependency struct {
	parent Dependency
}

func (c *commonDependency) GetParent() Dependency {
	return c.parent
}

// FIXME - janhajek move back cmd?
var Project *project
var Service *service

func init() {
	Project = &project{}
	Service = &service{
		commonDependency: commonDependency{
			parent: Project,
		},
	}
}

func getDependencyListFromRoot(dep Dependency) []Dependency {
	var list []Dependency
	for {
		if dep == nil {
			break
		}
		list = append(list, dep)
		dep = dep.GetParent()
	}

	slices.Reverse(list)

	return list
}
