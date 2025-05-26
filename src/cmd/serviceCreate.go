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
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/units"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
)

func serviceCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short("").
		ScopeLevel(cmdBuilder.ScopeProject()).
		StringFlag("zeropsYamlPath", "", i18n.T(i18n.ZeropsYamlLocation)).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("name", "", "Service name").
		StringFlag("mode", enumDefaultForFlag(enum.ServiceStackModeEnumNonHa), "Service mode "+enumValuesForFlag(enum.ServiceStackModeEnumAllPublic())).
		StringFlag("out", "", "Output format of command, using golang's text/template engine. Entity fields: "+formatAllowedTemplateFields(entity.ServiceFields)).
		StringFlag("envFile", "", "File with envs (will be set as secrets, runtime envs can be defined in zerops.yml)").
		StringSliceFlag("env", nil, "Envs to be set as secrets, runtime envs can be defined in zerops.yml. Accepts comma separated string or repeated flag. Format: {key}={value}").
		BoolFlag("startWithoutCode", false, "Start service immediately, empty without deploy").
		BoolFlag("noop", false, "Creates service only if none with the same name exists").
		HelpFlag("").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is nil")
			if err != nil {
				return err
			}

			startWithoutCode := cmdData.Params.GetBool("startWithoutCode")

			mode := cmdData.Params.GetString("mode")
			mode = strings.ToUpper(mode)
			if !enum.ServiceStackModeEnum(mode).Is(enum.ServiceStackModeEnumAllPublic()...) {
				return errors.Errorf("Invalid --mode, expected one of %s, got %s", enum.ServiceStackModeEnumAllPublic(), mode)
			}

			envFilePath := cmdData.Params.GetString("envFile")
			var envFile types.TextNull
			if envFilePath != "" {
				workingDir := cmdData.Params.GetString("workingDir")
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
					return errors.New("Env file must be a file")
				}
				const maxEnvFileSize = units.MiB
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
				cmdData.Params.GetString("workingDir"),
				cmdData.Params.GetString("zeropsYamlPath"),
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
				name, err = uxBlock.RunR(
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

			response, err := cmdData.RestApiClient.PostServiceStack(
				ctx,
				dtoPath.ServiceStackServiceStackTypeVersionId{ServiceStackTypeVersionId: "runtime"},
				body.PostStandardServiceStack{
					ProjectId:        project.ID,
					Name:             types.NewString(name),
					Mode:             gn.Ptr(enum.ServiceStackModeEnum(mode)),
					UserDataEnvFile:  envFile,
					StartWithoutCode: types.NewBoolNull(startWithoutCode),
				},
			)
			if err != nil {
				return err
			}

			noop := cmdData.Params.GetBool("noop")
			serviceStackProcess, err := response.Output()
			if err != nil {
				if apiError.HasErrorCode(err, errorCode.ServiceStackNameUnavailable) && noop {
					service, err := repository.GetServiceByName(ctx, cmdData.RestApiClient, project.ID, types.NewString(name))
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
						F:                   uxHelpers.CheckZeropsProcess(serviceStackProcess.Process.Id, cmdData.RestApiClient),
						RunningMessage:      "Creating service",
						ErrorMessageMessage: "Service creation failed",
						SuccessMessage:      "Service created",
					},
				},
			); err != nil {
				return err
			}

			service, err := repository.GetServiceById(ctx, cmdData.RestApiClient, serviceStackProcess.Id)
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
