package kongCmd

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// Auth is the authenticated handle. Held on CmdData.Auth, nil for guest.
// Require* methods perform scope resolution lazily — calling them is the
// only legitimate way for a command to obtain a Project / Service.
//
// All Require methods cache their result on the handle, so calling
// RequireProject twice in one command is free. RequireService transparently
// invokes RequireProject when it needs the project for name lookup.
type Auth struct {
	RestApiClient *zeropsRestApiClient.Handler
	VpnKeys       map[uuid.ProjectId]entity.VpnKey

	parent *CmdData

	project        optional.Null[entity.Project]
	service        optional.Null[entity.Service]
	projectPrinted bool
	servicePrinted bool
}

func newAuth(d *CmdData, client *zeropsRestApiClient.Handler, vpnKeys map[uuid.ProjectId]entity.VpnKey) *Auth {
	return &Auth{
		RestApiClient: client,
		VpnKeys:       vpnKeys,
		parent:        d,
	}
}

// RequireProject resolves the active project. Lookup order:
//  1. cached result
//  2. service already resolved → derive project from it
//  3. persisted scope (CliStorage.ScopeProjectId)
//  4. scope.ProjectID flag
//  5. interactive selector (+ create-new flow if WithCreateNewProject)
//
// Phase 3 fills in the implementation when the first project-scoped
// command is migrated. Until then this is intentionally a stub so the
// harness compiles without depending on logic that still lives in cmdBuilder.
func (a *Auth) RequireProject(_ context.Context, _ ProjectScope, _ ...ProjectOpt) (entity.Project, error) {
	if p, ok := a.project.Get(); ok {
		return p, nil
	}
	return entity.Project{}, errNotImplemented("Auth.RequireProject")
}

// RequireService resolves the active service. Lookup order:
//  1. cached result
//  2. scope.ServiceID flag → fetch by ID, project follows
//  3. scope.ServiceArg as UUID → fetch by ID, project follows
//  4. RequireProject + scope.ServiceArg as name → fetch by name
//  5. interactive selector (+ create-new flow if WithCreateNewService)
//
// Phase 3 fills in the implementation.
func (a *Auth) RequireService(_ context.Context, _ ServiceScope, _ ...ServiceOpt) (entity.Service, error) {
	if s, ok := a.service.Get(); ok {
		return s, nil
	}
	return entity.Service{}, errNotImplemented("Auth.RequireService")
}

// SelectProject opens the interactive project selector and returns the
// chosen project (or false if no selection was made). Used by `zcli scope
// project` and by RequireProject's interactive branch.
//
// Phase 3 fills in the implementation.
func (a *Auth) SelectProject(_ context.Context, _ ...uxHelpers.ProjectSelectorOption) (entity.Project, bool, error) {
	return entity.Project{}, false, errNotImplemented("Auth.SelectProject")
}

func errNotImplemented(name string) error {
	return notImplementedError{name: name}
}

type notImplementedError struct{ name string }

func (e notImplementedError) Error() string {
	return e.name + ": not yet implemented in kongCmd harness (Phase 3 stub)"
}
