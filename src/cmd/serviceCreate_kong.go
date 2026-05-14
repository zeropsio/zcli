package cmd

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/units"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
)

// ServiceCreateCmd creates a new service under the scoped project.
// ScopeProject (not Service) — the service is the thing we're creating.
type ServiceCreateCmd struct {
	kongCmd.ProjectScope

	ZeropsYamlPath   string   `name:"zerops-yaml-path" help:"Path to zerops.yaml (overrides working-dir lookup)."`
	WorkingDir       string   `name:"working-dir" default:"./" help:"Working directory for relative paths."`
	Name             string   `name:"name" help:"Service name."`
	Mode             string   `name:"mode" default:"NON_HA" help:"Service mode (NON_HA|HA)."`
	Out              string   `name:"out" help:"Output format using Go's text/template engine."`
	EnvFile          string   `name:"env-file" help:"File with secret envs."`
	EnvIsolation     string   `name:"env-isolation" default:"service" help:"Env isolation rule [service, none]."`
	SshIsolation     string   `name:"ssh-isolation" default:"vpn" help:"SSH isolation rule."`
	Location         string   `name:"location" hidden:"" help:"Service location (e.g. eu-central)."`
	Env              []string `name:"env" help:"Envs as secrets. {key}={value}, comma-separated or repeated."`
	StartWithoutCode bool     `name:"start-without-code" help:"Start service immediately, empty without deploy."`
	Noop             bool     `name:"noop" help:"Creates service only if none with the same name exists."`
}

func (c *ServiceCreateCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	mode := strings.ToUpper(c.Mode)
	if !enum.ServiceStackModeEnum(mode).Is(enum.ServiceStackModeEnumAllPublic()...) {
		return errors.Errorf("Invalid --mode, expected one of %s, got %s",
			enum.ServiceStackModeEnumAllPublic(), mode)
	}

	envFile, err := c.readEnvFile()
	if err != nil {
		return err
	}

	var outTemplate *template.Template
	if c.Out != "" {
		outTemplate, err = template.New("out").Parse(c.Out)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	configContent, err := yamlReader.ReadZeropsYamlContent(
		d.UxBlocks,
		c.WorkingDir,
		c.ZeropsYamlPath,
		yamlReader.WithReturnErrOnZeropsYamlNotFound(false),
	)
	if err != nil {
		return err
	}
	var suggestions []string
	if len(configContent) > 0 {
		setups, err := yamlReader.ReadZeropsYamlSetups(configContent)
		if err != nil {
			return err
		}
		suggestions = setups
	}

	name, err := c.resolveName(ctx, suggestions)
	if err != nil {
		return err
	}

	postService := entity.PostService{
		ProjectId:        project.Id,
		Name:             types.NewString(name),
		Mode:             enum.ServiceStackModeEnum(mode),
		EnvFile:          envFile,
		StartWithoutCode: types.NewBool(c.StartWithoutCode),
		SshIsolation:     types.NewStringNull(c.SshIsolation),
		EnvIsolation:     types.NewStringNull(c.EnvIsolation),
	}
	if c.Location != "" {
		postService.Location = stringId.NewLocationIdNullFromString(c.Location)
	}

	process, service, err := repository.PostGenericService(ctx, auth.RestApiClient, postService)
	if err != nil {
		if apiError.HasErrorCode(err, errorCode.ServiceStackNameUnavailable) && c.Noop {
			existing, getErr := repository.GetServiceByName(ctx, auth.RestApiClient, project.Id, types.NewString(name))
			if getErr != nil {
				return getErr
			}
			d.UxBlocks.PrintInfoText("Service with the same name already exists")
			if outTemplate != nil {
				if err := outTemplate.Execute(d.Stdout, existing); err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		}
		return err
	}

	if err := uxHelpers.ProcessCheckWithSpinner(ctx, d.UxBlocks, []uxHelpers.Process{{
		F:                   uxHelpers.CheckZeropsProcess(process.Id, auth.RestApiClient, d.UxBlocks),
		RunningMessage:      "Creating service",
		ErrorMessageMessage: "Service creation failed",
		SuccessMessage:      "Service created",
	}}); err != nil {
		return err
	}

	service, err = repository.GetServiceById(ctx, auth.RestApiClient, service.Id)
	if err != nil {
		return err
	}
	if outTemplate != nil {
		if err := outTemplate.Execute(d.Stdout, service); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (c *ServiceCreateCmd) readEnvFile() (types.TextNull, error) {
	var envFile types.TextNull
	if c.EnvFile != "" {
		absPath, err := filepath.Abs(path.Join(c.WorkingDir, c.EnvFile))
		if err != nil {
			return envFile, errors.WithStack(err)
		}
		stat, err := os.Stat(absPath)
		if err != nil {
			return envFile, errors.WithStack(err)
		}
		if stat.IsDir() {
			return envFile, errors.New("--env-file must point to a file")
		}
		if stat.Size() > int64(maxEnvFileSize) {
			return envFile, errors.Errorf("Env file size too large, max allowed size %s",
				units.ByteCountIEC(maxEnvFileSize))
		}
		content, err := os.ReadFile(absPath)
		if err != nil {
			return envFile, errors.WithStack(err)
		}
		envFile = types.NewTextNull(string(content))
	}
	if len(c.Env) > 0 {
		envs := strings.Join(c.Env, "\n")
		if existing, filled := envFile.Get(); filled {
			envFile = types.NewTextNull(existing.Native() + "\n" + envs)
		} else {
			envFile = types.NewTextNull(envs)
		}
	}
	return envFile, nil
}

func (c *ServiceCreateCmd) resolveName(ctx context.Context, suggestions []string) (string, error) {
	if c.Name != "" {
		return c.Name, nil
	}
	if !terminal.IsTerminal() {
		return "", errors.New("Must specify name with --name")
	}
	label := styles.NewStringBuilder()
	label.WriteString("Type ")
	label.WriteStyledString(styles.SelectStyle().Bold(true), "service")
	label.WriteString(" name")
	return uxBlock.Run(
		input.NewRoot(
			ctx,
			input.WithLabel(label.String()),
			input.WithHelpPlaceholder(),
			input.WithPlaceholderStyle(styles.HelpStyle()),
			input.WithoutPrompt(),
			input.WithSetSuggestions(suggestions),
		),
		input.GetValueFunc,
	)
}
