package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

// resolveScopedOrg resolves the org an org-level command (project create / import) should act on,
// in precedence order:
//
//  1. the explicit --org-id flag
//  2. the configured client scope - repo profile override, else global default - via Profile.ScopeClient
//  3. the only org on the account, picked silently
//  4. the interactive org selector (terminal only; non-terminal with no flag/scope/single-org errors)
//
// Steps 1, 3 and 4 are the pre-existing behaviour shared by the two commands; step 2 slots the new
// client scope in just below the explicit flag.
func resolveScopedOrg(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) (entity.Org, error) {
	if flag := cmdData.Params.GetString("org-id"); flag != "" {
		return repository.GetOrgById(ctx, cmdData.RestApiClient, entity.OrgId(flag))
	}

	if id, ok := cmdData.Profile.ScopeClient(); ok {
		return repository.GetOrgById(ctx, cmdData.RestApiClient, id)
	}

	orgs, err := repository.GetAllOrgs(ctx, cmdData.RestApiClient)
	if err != nil {
		return entity.Org{}, err
	}
	if len(orgs) == 1 {
		return orgs[0], nil
	}
	if !terminal.IsTerminal() {
		return entity.Org{}, errors.New("Must specify organization ID with --org-id")
	}
	return uxHelpers.PrintOrgSelector(ctx, cmdData.RestApiClient, uxHelpers.WithOrgPickOnlyOneItem(true))
}
