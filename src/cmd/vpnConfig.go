package cmd

import (
	"context"
	"os"
	"time"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func vpnConfigCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("config").
		Short("Generate VPN configuration file without connecting").
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		IntFlag(vpnFlagMtu, 1420, i18n.T(i18n.VpnMtuFlag)).
		BoolFlag(vpnFlagSkipDnsSetup, false, "skip DNS configuration - you will need to use IP addresses to connect to services instead of domain names").
		StringFlag(vpnFlagOutput, "", "output file path (use '-' for stdout, empty for default location)").
		HelpFlag("Generate WireGuard VPN configuration file for the project without establishing connection").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			dnsSetup := !cmdData.Params.GetBool(vpnFlagSkipDnsSetup)

			uxBlocks := cmdData.UxBlocks
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			privateKey, err := getOrCreatePrivateVpnKey(project, cmdData)
			if err != nil {
				return err
			}

			publicKey := privateKey.PublicKey()

			postProjectResponse, err := cmdData.RestApiClient.PostProjectVpn(
				ctx,
				path.ProjectId{Id: project.Id},
				body.PostProjectVpn{PublicKey: types.String(publicKey.String())},
			)
			if err != nil {
				return err
			}

			vpnSettings, err := postProjectResponse.Output()
			if err != nil {
				return err
			}

			outputPath := cmdData.Params.GetString(vpnFlagOutput)

			// Determine output destination
			var f *os.File
			var filePath string
			var fileMode os.FileMode

			switch outputPath {
			case "-":
				// Output to stdout
				f = os.Stdout
				filePath = "stdout"
			case "":
				// Use default location
				filePath, fileMode, err = constants.WgConfigFilePath()
				if err != nil {
					return err
				}
				f, err = file.Open(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
				if err != nil {
					return err
				}
				defer f.Close()
			default:
				// Use custom file path
				filePath = outputPath
				fileMode = 0600
				f, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
				if err != nil {
					return err
				}
				defer f.Close()
			}

			if err := wg.GenerateConfig(f, privateKey, vpnSettings, cmdData.Params.GetInt(vpnFlagMtu), dnsSetup); err != nil {
				return err
			}

			if outputPath != "-" {
				uxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.VpnConfigSaved), filePath))
			}

			if _, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				if data.ProjectVpnKeyRegistry == nil {
					data.ProjectVpnKeyRegistry = make(map[uuid.ProjectId]entity.VpnKey)
				}
				data.ProjectVpnKeyRegistry[project.Id] = entity.VpnKey{
					ProjectId: project.Id,
					Key:       privateKey.String(),
					CreatedAt: time.Now(),
				}

				return data
			}); err != nil {
				return err
			}

			return nil
		})
}
