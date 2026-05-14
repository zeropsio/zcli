package kongCmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/apiError"
)

// RunOptions controls how RunRootCmd executes. Zero value matches production
// defaults used by ExecuteRootCmd. Shape mirrors cmdBuilder.RunOptions so the
// integration test fixture migrates with a one-line import change.
type RunOptions struct {
	//nolint:containedctx // intentional: callers (tests, ExecuteRootCmd) pass the root ctx by value
	Ctx    context.Context
	Args   []string
	Stdout io.Writer
	Stderr io.Writer
}

// ExecuteRootCmd is the production entry point — runs with default options
// and exits the process with the resulting exit code.
func ExecuteRootCmd(cli any) {
	os.Exit(RunRootCmd(cli, RunOptions{}))
}

// RunRootCmd is the test-friendly entry point. cli is a pointer to the kong
// grammar struct. Never calls os.Exit; returns the exit code instead.
func RunRootCmd(cli any, opts RunOptions) int {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	ctx := opts.Ctx
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithCancel(context.Background())
		regSignals(cancel)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()
	ctx = support.Context(ctx)

	isTerminal := terminal.IsTerminal()
	termWidth, termHeight, _ := term.GetSize(0)
	outputLogger, debugFileLogger := createLoggers(isTerminal, stderr)
	uxBlocks := uxBlock.NewBlocks(outputLogger, debugFileLogger, isTerminal, termWidth, termHeight, cancel)

	store, err := createCliStorage()
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	cfg := LoadConfig(stderr)

	d := &CmdData{
		CliStorage:        store,
		UxBlocks:          uxBlocks,
		Stdout:            printer.NewPrinter(stdout),
		Stderr:            printer.NewPrinter(stderr),
		Config:            cfg,
		LocalZcliYamlFile: localFileOrEmpty(cfg),
	}

	token := store.Data().Token
	if env := os.Getenv(constants.CliTokenEnvVar); env != "" {
		token = env
	}
	if token != "" {
		host := store.Data().RegionData.Address
		if host == "" {
			host = constants.DefaultRegion
		}
		d.Auth = newAuth(d, zeropsRestApiClient.NewAuthorizedClient(token, host), store.Data().ProjectVpnKeyRegistry)
	}

	parser, err := kong.New(cli,
		kong.Name("zcli"),
		kong.Resolvers(cfg),
		kong.Writers(stdout, stderr),
		kong.Exit(func(int) {}), // we own the exit code via the int return
	)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	args := opts.Args
	if args == nil {
		args = os.Args[1:]
	}
	kctx, err := parser.Parse(args)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	leaf := kctx.Selected()
	if leaf == nil {
		// no subcommand: kong prints help by default — nothing to dispatch
		return 0
	}

	runner, ok := leaf.Target.Addr().Interface().(Runner)
	if !ok {
		return errorExitCode(errors.Errorf("command %q has no Run method", kctx.Command()), uxBlocks)
	}
	if err := runner.Run(ctx, d); err != nil {
		return errorExitCode(err, uxBlocks)
	}
	return 0
}

func localFileOrEmpty(c *Config) string {
	if p, ok := c.LocalZcliYamlFile(); ok {
		return p
	}
	return ""
}

func errorExitCode(err error, uxBlocks uxBlock.UxBlocks) int {
	if err == nil {
		return 0
	}
	uxBlocks.LogDebug(fmt.Sprintf("error: %+v", err))

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintErrorText(err.Error())
		return 1
	}

	var apiErr apiError.Error
	if errors.As(err, &apiErr) {
		uxBlocks.PrintErrorText(apiErr.GetMessage())
		if apiErr.GetMeta() != nil {
			meta, err := yaml.Marshal(apiErr.GetMeta())
			if err != nil {
				uxBlocks.PrintErrorText(fmt.Sprintf("couldn't parse meta of error: %s", apiErr.GetMessage()))
			}
			uxBlocks.PrintErrorText(string(meta))
		}
		return 1
	}

	if errors.Is(err, models.ErrCtrlC) {
		uxBlocks.PrintInfo(styles.InfoLine("canceled"))
		return 0
	}

	uxBlocks.PrintErrorText(err.Error())
	return 1
}

func createLoggers(isTerminal bool, stderr io.Writer) (*logger.Handler, *logger.Handler) {
	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
		Out:        stderr,
	})
	loggerFilePath, fileMode, err := constants.LogFilePath()
	if err != nil {
		outputLogger.Warning(styles.WarningLine(err.Error()))
	}
	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: loggerFilePath,
		FileMode: fileMode,
	})
	return outputLogger, debugFileLogger
}

func regSignals(cancel func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
}

func createCliStorage() (*cliStorage.Handler, error) {
	filePath, fileMode, err := constants.CliDataFilePath()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[cliStorage.Data](storage.Config{
		FilePath: filePath,
		FileMode: fileMode,
	})
	return &cliStorage.Handler{Handler: s}, err
}
