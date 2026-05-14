package kongCmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// Auth is the authenticated handle. Held on CmdData.Auth, nil for guest.
// Require* methods perform scope resolution lazily — calling them is the
// only legitimate way for a command to obtain a Project / Service.
//
// All Require methods cache their result on the handle: calling
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

// RequireProject resolves the active project. Lookup order, first hit wins:
//  1. cached
//  2. service already resolved → derive project from it
//  3. persisted scope (CliStorage.ScopeProjectId); resets the scope on
//     ProjectNotFound, propagates other errors
//  4. scope.ProjectID (the --project-id flag, or whatever the command put
//     there from its own positional)
//  5. interactive selector + optional create-new flow (WithCreateNewProject)
//
// Side effect: "Selected project: <name>" is printed once per call site
// (idempotent across repeated Requires within a command).
func (a *Auth) RequireProject(ctx context.Context, scope ProjectScope, opts ...ProjectOpt) (entity.Project, error) {
	if p, ok := a.project.Get(); ok {
		return p, nil
	}

	cfg := projectReqCfg{}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.skipSelect {
		return entity.Project{}, nil
	}

	// 1. service already known → derive project from it
	if svc, ok := a.service.Get(); ok {
		if project, err := repository.GetProjectById(ctx, a.RestApiClient, svc.ProjectId); err == nil {
			a.setProject(project)
			a.printProjectOnce()
			return project, nil
		}
	}

	// 2. persisted scope from cliStorage
	if a.parent.CliStorage.Data().ScopeProjectId.Filled() {
		id, _ := a.parent.CliStorage.Data().ScopeProjectId.Get()
		project, err := repository.GetProjectById(ctx, a.RestApiClient, id)
		if err != nil {
			if errorsx.Is(err, errorsx.ErrorCode(errorCode.ProjectNotFound)) {
				if err := projectScopeReset(a.parent.CliStorage); err != nil {
					return entity.Project{}, err
				}
			} else {
				return entity.Project{}, err
			}
		} else {
			a.setProject(project)
		}
	}

	// 3. explicit identifier (--project-id flag value supplied by the command)
	if scope.ProjectID != "" && !a.project.Filled() {
		project, err := repository.GetProjectById(ctx, a.RestApiClient, uuid.ProjectId(scope.ProjectID))
		if err != nil {
			return entity.Project{}, errorsx.Convert(err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ apiError.Error, meta map[string]interface{}) string {
							return i18n.T(i18n.ErrorInvalidProjectId, scope.ProjectID, meta["message"])
						},
					),
				),
			)
		}
		a.setProject(project)
	}

	// 4. interactive (+ create-new) selector
	if !a.project.Filled() {
		project, selected, err := a.SelectProject(ctx, uxHelpers.WithCreateNewProject(cfg.createNew))
		if err != nil {
			return entity.Project{}, err
		}
		if selected {
			a.setProject(project)
		}
	}

	a.printProjectOnce()
	if p, ok := a.project.Get(); ok {
		return p, nil
	}
	return entity.Project{}, nil
}

// RequireService resolves the active service. Lookup order:
//  1. cached
//  2. scope.ServiceID flag → fetch by ID, project follows from it
//  3. scope.ServiceArg as UUID (22 chars) → fetch by ID, project follows
//  4. RequireProject + scope.ServiceArg as name → fetch by name
//  5. interactive selector (create-new path not yet implemented; serviceDeploy
//     and servicePush will add it when they migrate)
//
// "Selected service: <name>" prints once per Auth lifetime.
func (a *Auth) RequireService(ctx context.Context, scope ServiceScope, opts ...ServiceOpt) (entity.Service, error) {
	if s, ok := a.service.Get(); ok {
		return s, nil
	}
	cfg := serviceReqCfg{}
	for _, o := range opts {
		o(&cfg)
	}

	// 1. --service-id flag wins outright
	if scope.ServiceID != "" {
		svc, err := repository.GetServiceById(ctx, a.RestApiClient, uuid.ServiceStackId(scope.ServiceID))
		if err != nil {
			return entity.Service{}, errorsx.Convert(err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ apiError.Error, meta map[string]interface{}) string {
							return i18n.T(i18n.ErrorInvalidServiceId, scope.ServiceID, meta["message"])
						},
					),
				),
			)
		}
		a.setService(svc)
		// derive project so subsequent RequireProject in this run is free
		if !a.project.Filled() {
			if project, err := repository.GetProjectById(ctx, a.RestApiClient, svc.ProjectId); err == nil {
				a.setProject(project)
			}
		}
		a.printServiceOnce()
		return svc, nil
	}

	// 2. positional that looks like a UUID
	if len(scope.ServiceArg) == 22 {
		if svc, err := repository.GetServiceById(ctx, a.RestApiClient, uuid.ServiceStackId(scope.ServiceArg)); err == nil {
			a.setService(svc)
			if !a.project.Filled() {
				if project, err := repository.GetProjectById(ctx, a.RestApiClient, svc.ProjectId); err == nil {
					a.setProject(project)
				}
			}
			a.printServiceOnce()
			return svc, nil
		}
	}

	// 3. resolve project then look up by name
	project, err := a.RequireProject(ctx, scope.ProjectScope)
	if err != nil {
		return entity.Service{}, err
	}

	if scope.ServiceArg != "" {
		svc, err := repository.GetServiceByIdOrName(ctx, a.RestApiClient, project.Id, scope.ServiceArg)
		if err != nil {
			return entity.Service{}, err
		}
		a.setService(svc)
		a.printServiceOnce()
		return svc, nil
	}

	// 4. interactive selector
	selected, err := uxHelpers.PrintServiceSelector(
		ctx,
		a.RestApiClient,
		project,
		uxHelpers.WithCreateNewService(cfg.createNew),
	)
	if err != nil {
		return entity.Service{}, err
	}
	if selected.Filled() {
		a.setService(selected.Some())
	} else if cfg.createNew && terminal.IsTerminal() {
		svc, err := a.createNewService(ctx, project, cfg.workingDir, cfg.zeropsYamlPath)
		if err != nil {
			return entity.Service{}, err
		}
		a.setService(svc)
	}

	a.printServiceOnce()
	if s, ok := a.service.Get(); ok {
		return s, nil
	}
	return entity.Service{}, nil
}

// SelectProject runs the interactive picker and, if no project is chosen and
// the process is TTY-attached, the create-new flow. Returns ({}, false, nil)
// when nothing was selected and we cannot create (non-TTY).
func (a *Auth) SelectProject(ctx context.Context, opts ...uxHelpers.ProjectSelectorOption) (entity.Project, bool, error) {
	selected, err := uxHelpers.PrintProjectSelector(ctx, a.RestApiClient, opts...)
	if err != nil {
		return entity.Project{}, false, err
	}
	if selected.Filled() {
		return selected.Some(), true, nil
	}
	if terminal.IsTerminal() {
		project, err := a.createNewProject(ctx)
		if err != nil {
			return entity.Project{}, false, err
		}
		return project, true, nil
	}
	return entity.Project{}, false, nil
}

// createNewProject runs the org → name → location flow and POSTs a new
// Project to the API. Called from SelectProject when interactive and no
// project was picked from the existing list.
func (a *Auth) createNewProject(ctx context.Context) (entity.Project, error) {
	selectedOrg, err := uxHelpers.PrintOrgSelector(ctx, a.RestApiClient, uxHelpers.WithOrgPickOnlyOneItem(true))
	if err != nil {
		return entity.Project{}, err
	}
	a.parent.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected org", selectedOrg.Name.String()))

	label := styles.NewStringBuilder()
	label.WriteString("Type ")
	label.WriteStyledString(styles.SelectStyle().Bold(true), "project")
	label.WriteString(" name")

	name, err := uxBlock.Run(
		input.NewRoot(
			ctx,
			input.WithLabel(label.String()),
			input.WithHelpPlaceholder(),
			input.WithPlaceholderStyle(styles.HelpStyle()),
			input.WithoutPrompt(),
		),
		input.GetValueFunc,
	)
	if err != nil {
		return entity.Project{}, err
	}

	location, err := uxHelpers.PrintLocationSelector(ctx, a.RestApiClient)
	if err != nil {
		return entity.Project{}, err
	}
	a.parent.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected location", location.Name.String()))

	project, err := repository.PostProject(ctx, a.RestApiClient, entity.PostProject{
		OrgId:    selectedOrg.Id,
		Name:     types.NewString(name),
		Mode:     enum.ProjectModeEnumLight,
		Location: location.Id.LocationIdNull(),
	})
	if err != nil {
		return entity.Project{}, err
	}

	a.parent.UxBlocks.PrintSuccessText("Project created")
	return project, nil
}

// createNewService runs the interactive new-service flow: read zerops.yaml
// for setup-name suggestions, prompt the user, POST a fresh service to the
// project, poll the resulting Process, and return the refreshed Service.
// Caller is responsible for setting it on Auth.
func (a *Auth) createNewService(ctx context.Context, project entity.Project, workingDir, zeropsYamlPath string) (entity.Service, error) {
	configContent, err := yamlReader.ReadZeropsYamlContent(a.parent.UxBlocks, workingDir, zeropsYamlPath)
	if err != nil {
		return entity.Service{}, err
	}
	setups, err := yamlReader.ReadZeropsYamlSetups(configContent)
	if err != nil {
		return entity.Service{}, err
	}

	label := styles.NewStringBuilder()
	label.WriteString("Type ")
	label.WriteStyledString(styles.SelectStyle().Bold(true), "service")
	label.WriteString(" name")

	name, err := uxBlock.Run(
		input.NewRoot(
			ctx,
			input.WithLabel(label.String()),
			input.WithHelpPlaceholder(),
			input.WithPlaceholderStyle(styles.HelpStyle()),
			input.WithoutPrompt(),
			input.WithSetSuggestions(setups),
		),
		input.GetValueFunc,
	)
	if err != nil {
		return entity.Service{}, err
	}

	process, service, err := repository.PostGenericService(ctx, a.RestApiClient, entity.PostService{
		ProjectId: project.Id,
		Name:      types.NewString(name),
		Mode:      enum.ServiceStackModeEnumNonHa,
	})
	if err != nil {
		if apiError.HasErrorCode(err, errorCode.ServiceStackNameUnavailable) {
			existing, err := repository.GetServiceByName(ctx, a.RestApiClient, project.Id, types.NewString(name))
			if err != nil {
				return entity.Service{}, err
			}
			return existing, nil
		}
		return entity.Service{}, err
	}

	if err := uxHelpers.ProcessCheckWithSpinner(
		ctx,
		a.parent.UxBlocks,
		[]uxHelpers.Process{{
			F:                   uxHelpers.CheckZeropsProcess(process.Id, a.RestApiClient, a.parent.UxBlocks),
			RunningMessage:      "Creating service",
			ErrorMessageMessage: "Service creation failed",
			SuccessMessage:      "Service created",
		}},
	); err != nil {
		return entity.Service{}, err
	}
	return repository.GetServiceById(ctx, a.RestApiClient, service.Id)
}

func (a *Auth) setProject(p entity.Project) { a.project = optional.New(p) }
func (a *Auth) setService(s entity.Service) { a.service = optional.New(s) }

func (a *Auth) printServiceOnce() {
	if a.servicePrinted {
		return
	}
	if s, ok := a.service.Get(); ok {
		a.parent.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.SelectedService), s.Name.String()))
	}
	a.servicePrinted = true
}

func (a *Auth) printProjectOnce() {
	if a.projectPrinted {
		return
	}
	if p, ok := a.project.Get(); ok {
		a.parent.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.SelectedProject), p.Name.String()))
	}
	a.projectPrinted = true
}

func projectScopeReset(store *cliStorage.Handler) error {
	_, err := store.Update(func(d cliStorage.Data) cliStorage.Data {
		d.ScopeProjectId = uuid.ProjectIdNull{}
		return d
	})
	return err
}

func errNotImplemented(name string) error {
	return notImplementedError{name: name}
}

type notImplementedError struct{ name string }

func (e notImplementedError) Error() string {
	return e.name + ": not yet implemented in kongCmd harness (phase 3 stub)"
}
