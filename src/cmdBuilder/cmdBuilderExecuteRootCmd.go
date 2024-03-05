package cmdBuilder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/apiError"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

func (b *CmdBuilder) CreateAndExecuteRootCobraCmd() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	regSignals(cancel)
	ctx = support.Context(ctx)

	isTerminal := isTerminal()

	width, _, err := term.GetSize(0)
	if err != nil {
		width = 100
	}

	outputLogger, debugFileLogger := createLoggers(isTerminal)

	uxBlocks := uxBlock.NewBlock(outputLogger, debugFileLogger, isTerminal, width, cancel)

	defer func() {
		if err != nil {
			printError(err, uxBlocks)
		}
	}()

	cliStorage, err := createCliStorage()
	if err != nil {
		return err
	}

	flagParams := flagParams.New()

	rootCmd := createRootCommand()

	for _, cmd := range b.commands {
		cobraCmd, err := b.buildCobraCmd(cmd, flagParams, uxBlocks, cliStorage)
		if err != nil {
			return err
		}
		rootCmd.AddCommand(cobraCmd)
	}

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		printError(err, uxBlocks)
	}

	return nil
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "zcli",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		SilenceErrors:     true,
	}

	rootCmd.SetHelpTemplate(`` + styles.CobraSectionColor().SetString("Usage:").String() + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + styles.CobraSectionColor().SetString("Aliases:").String() + `
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + styles.CobraSectionColor().SetString("Examples:").String() + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

` + styles.CobraSectionColor().SetString("Available Commands:").String() + `{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

` + styles.CobraSectionColor().SetString("Additional Commands:").String() + `{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + styles.CobraSectionColor().SetString("Flags:").String() + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + styles.CobraSectionColor().SetString("Global Flags:").String() + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + styles.CobraSectionColor().SetString("Additional help topics:").String() + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .CommandPath .CommandPathPadding}}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

` + styles.CobraSectionColor().SetString("Global Env Variables:").String() + `
  ` + styles.CobraItemNameColor().SetString(constants.CliLogFilePathEnvVar).String() + `     ` + i18n.T(i18n.CliLogFilePathEnvVar) + `
  ` + styles.CobraItemNameColor().SetString(constants.CliDataFilePathEnvVar).String() + `    ` + i18n.T(i18n.CliDataFilePathEnvVar) + `
  ` + styles.CobraItemNameColor().SetString(constants.CliTerminalMode).String() + `     ` + i18n.T(i18n.CliTerminalModeEnvVar) + `

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	return rootCmd
}

func printError(err error, uxBlocks uxBlock.UxBlocks) {
	uxBlocks.LogDebug(fmt.Sprintf("error: %+v", err))

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintError(styles.ErrorLine(err.Error()))
		return
	}

	var apiErr apiError.Error
	if errors.As(err, &apiErr) {
		uxBlocks.PrintError(styles.ErrorLine(apiErr.GetMessage()))
		if apiErr.GetMeta() != nil {
			meta, err := yaml.Marshal(apiErr.GetMeta())
			if err != nil {
				uxBlocks.PrintError(styles.ErrorLine(fmt.Sprintf("couldn't parse meta of error: %s", apiErr.GetMessage())))
			}
			uxBlocks.PrintError(styles.ErrorLine(string(meta)))
		}

		return
	}

	uxBlocks.PrintError(styles.ErrorLine(err.Error()))
}

type terminalMode string

const (
	TerminalModeAuto     terminalMode = "auto"
	TerminalModeDisabled terminalMode = "disabled"
	TerminalModeEnabled  terminalMode = "enabled"
)

func isTerminal() bool {
	env := os.Getenv(constants.CliTerminalMode)

	switch terminalMode(env) {
	case TerminalModeAuto, "":
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	case TerminalModeDisabled:
		return false
	case TerminalModeEnabled:
		return true
	default:
		os.Stdout.WriteString(styles.WarningLine(i18n.T(i18n.UnknownTerminalMode, env)).String())

		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}
}

func createLoggers(isTerminal bool) (*logger.Handler, *logger.Handler) {
	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
	})

	loggerFilePath, err := constants.LogFilePath()
	if err != nil {
		outputLogger.Warning(styles.WarningLine(err.Error()))
	}

	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: loggerFilePath,
	})

	return outputLogger, debugFileLogger
}

func regSignals(contextCancel func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		contextCancel()
	}()
}

func createCliStorage() (*cliStorage.Handler, error) {
	filePath, err := constants.CliDataFilePath()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[cliStorage.Data](
		storage.Config{
			FilePath: filePath,
		},
	)
	return &cliStorage.Handler{Handler: s}, err
}
