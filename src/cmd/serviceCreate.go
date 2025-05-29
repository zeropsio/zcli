package cmd

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
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
)

const maxEnvFileSize = units.MiB

func serviceCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short("Crates a new project for specified project.").
		ScopeLevel(cmdBuilder.ScopeProject()).
		StringFlag("zerops-yaml-path", "", i18n.T(i18n.ZeropsYamlLocation)).
		StringFlag("working-dir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("name", "", "Service name").
		StringFlag("mode", enumDefaultForFlag(enum.ServiceStackModeEnumNonHa), "Service mode "+enumValuesForFlag(enum.ServiceStackModeEnumAllPublic())).
		StringFlag("out", "", "Output format of command, using golang's text/template engine. Entity fields: "+formatAllowedTemplateFields(entity.ServiceFields)).
		StringFlag("env-file", "", "File with envs (will be set as secrets, runtime envs can be defined in zerops.yml). Max file size is "+units.ByteCountIEC(maxEnvFileSize)).
		StringFlag("env-isolation", "service", "Env isolation rule [service, none] for more info see docs https://docs.zerops.io/features/env-variables#isolation-modes").
		StringFlag("ssh-isolation", "vpn", "SSH isolation rules, for more info see docs https://docs.zerops.io/references/ssh#ssh-access-control").
		StringSliceFlag("env", nil, "Envs to be set as secrets, runtime envs can be defined in zerops.yml. Accepts comma separated string or repeated flag. Format: {key}={value}").
		BoolFlag("start-without-code", false, "Start service immediately, empty without deploy").
		BoolFlag("noop", false, "Creates service only if none with the same name exists").
		HelpFlag("Help for the service create command.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is nil")
			if err != nil {
				return err
			}

			startWithoutCode := cmdData.Params.GetBool("start-without-code")

			mode := cmdData.Params.GetString("mode")
			mode = strings.ToUpper(mode)
			if !enum.ServiceStackModeEnum(mode).Is(enum.ServiceStackModeEnumAllPublic()...) {
				return errors.Errorf("Invalid --mode, expected one of %s, got %s", enum.ServiceStackModeEnumAllPublic(), mode)
			}

			envFilePath := cmdData.Params.GetString("env-file")
			var envFile types.TextNull
			if envFilePath != "" {
				workingDir := cmdData.Params.GetString("working-dir")
				envFilePath = path.Join(workingDir, envFilePath)
				envFilePath, err = filepath.Abs(envFilePath)
				if err != nil {
					return errors.WithStack(err)
				}
				stat, err := os.Stat(envFilePath)
				if err != nil {
					return errors.WithStack(err)
				}
				if stat.IsDir() {
					return errors.New("--env-file must point to a file")
				}
				if stat.Size() > int64(maxEnvFileSize) {
					return errors.Errorf("Env file size too large, max allowed size %s", units.ByteCountIEC(maxEnvFileSize))
				}
				envFileContent, err := os.ReadFile(envFilePath)
				if err != nil {
					return errors.WithStack(err)
				}
				envFile = types.NewTextNull(string(envFileContent))
			}

			envSlice := cmdData.Params.GetStringSlice("env")
			if len(envSlice) > 0 {
				envs := strings.Join(envSlice, "\n")
				if f, filled := envFile.Get(); filled {
					envFile = types.NewTextNull(f.Native() + "\n" + envs)
				} else {
					envFile = types.NewTextNull(envs)
				}
			}

			outFormat := cmdData.Params.GetString("out")
			var outTemplate *template.Template
			if outFormat != "" {
				outTemplate, err = template.New("out").Parse(outFormat)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			configContent, err := yamlReader.ReadZeropsYamlContent(
				cmdData.UxBlocks,
				cmdData.Params.GetString("working-dir"),
				cmdData.Params.GetString("zerops-yaml-path"),
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

			label := styles.NewStringBuilder()
			label.WriteString("Type ")
			label.WriteStyledString(
				styles.SelectStyle().
					Bold(true),
				"service",
			)
			label.WriteString(" name")

			name := cmdData.Params.GetString("name")
			if name == "" && terminal.IsTerminal() {
				name, err = uxBlock.Run(
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
				if err != nil {
					return err
				}
			} else if name == "" {
				return errors.New("Must specify name with --name")
			}

			process, service, err := repository.PostGenericService(
				ctx,
				cmdData.RestApiClient,
				entity.PostService{
					ProjectId:        project.Id,
					Name:             types.NewString(name),
					Mode:             enum.ServiceStackModeEnum(mode),
					EnvFile:          envFile,
					StartWithoutCode: types.NewBool(startWithoutCode),
					SshIsolation:     types.NewStringNull(cmdData.Params.GetString("ssh-isolation")),
					EnvIsolation:     types.NewStringNull(cmdData.Params.GetString("env-isolation")),
				},
			)
			if err != nil {
				noop := cmdData.Params.GetBool("noop")
				if apiError.HasErrorCode(err, errorCode.ServiceStackNameUnavailable) && noop {
					service, err := repository.GetServiceByName(ctx, cmdData.RestApiClient, project.Id, types.NewString(name))
					if err != nil {
						return err
					}
					cmdData.UxBlocks.PrintInfoText("Service with the same name already exists")
					if outTemplate != nil {
						if err := outTemplate.Execute(cmdData.Stdout, service); err != nil {
							return errors.WithStack(err)
						}
					}
					return nil
				}
				return err
			}

			if err := uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{
					{
						F:                   uxHelpers.CheckZeropsProcess(process.Id, cmdData.RestApiClient),
						RunningMessage:      "Creating service",
						ErrorMessageMessage: "Service creation failed",
						SuccessMessage:      "Service created",
					},
				},
			); err != nil {
				return err
			}

			service, err = repository.GetServiceById(ctx, cmdData.RestApiClient, service.Id)
			if err != nil {
				return err
			}

			if outTemplate != nil {
				if err := outTemplate.Execute(cmdData.Stdout, service); err != nil {
					return errors.WithStack(err)
				}
			}

			return nil
		})
}
