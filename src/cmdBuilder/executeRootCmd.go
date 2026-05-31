package cmdBuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	bubblesSpinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// runOptions controls how RunRootCmd executes. Build via the With... helpers.
type runOptions struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

// RunOption configures a RunRootCmd invocation.
type RunOption func(*runOptions)

// WithArgs overrides os.Args[1:]. Useful for tests.
func WithArgs(args []string) RunOption {
	return func(o *runOptions) { o.args = args }
}

// WithStdout redirects command output. Defaults to os.Stdout.
func WithStdout(w io.Writer) RunOption {
	return func(o *runOptions) { o.stdout = w }
}

// WithStderr redirects error/log output (including uxBlock messages).
// Defaults to os.Stderr.
func WithStderr(w io.Writer) RunOption {
	return func(o *runOptions) { o.stderr = w }
}

func newRunOptions(opts ...RunOption) runOptions {
	o := runOptions{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

var matchFirstCap = regexp.MustCompile("([A-Z]+)")

func camelCaseToKebabCase(camel string) string {
	return strings.ToLower(matchFirstCap.ReplaceAllString(camel, "-${1}"))
}

func normalizeFlagNames(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(camelCaseToKebabCase(name))
}

// ExecuteRootCmd runs the CLI with production defaults and exits the process
// with the resulting status code.
func ExecuteRootCmd(rootCmd *Cmd) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	code := RunRootCmd(ctx, rootCmd)
	stop()
	os.Exit(code)
}

// RunRootCmd is the test-friendly entry point. It never calls os.Exit and
// instead returns the exit code the process should use.
func RunRootCmd(ctx context.Context, rootCmd *Cmd, opts ...RunOption) int {
	o := newRunOptions(opts...)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = support.Context(ctx)

	isTerminal := terminal.IsTerminal()
	terminalWidth, terminalHeight, _ := term.GetSize(0)
	outputLogger, debugFileLogger := createLoggers(isTerminal, o.stderr)

	uxBlocks := uxBlock.NewBlocks(outputLogger, debugFileLogger, isTerminal, terminalWidth, terminalHeight, cancel)

	profileManager, err := createProfile(ctx, uxBlocks, isTerminal)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	flagParams := flagParams.New(o.stderr)

	//nolint:contextcheck // cobra threads ctx into the run-func via ExecuteContext below; pulling from cobraCmd.Context() inside createCmdRunFunc is the canonical pattern.
	cobraCmd, err := buildCobraCmd(rootCmd, flagParams, uxBlocks, profileManager, o.stdout, o.stderr)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	cobraCmd.SetGlobalNormalizationFunc(normalizeFlagNames)
	cobraCmd.SetOut(o.stdout)
	cobraCmd.SetErr(o.stderr)

	// --profile picks a repo profile entry to act as if it were the active one for this
	// invocation. The persistent flag is declared on the root so it inherits everywhere; the
	// actual switch happens inside createCmdRunFunc once cobra has parsed the value.
	cobraCmd.PersistentFlags().String(RootProfileFlagName, "", "Override active repo profile for this invocation (must exist in .zcli/profile.json)")
	if o.args != nil {
		cobraCmd.SetArgs(o.args)
	}

	if err := cobraCmd.ExecuteContext(ctx); err != nil {
		return errorExitCode(err, uxBlocks)
	}
	return 0
}

func errorExitCode(err error, uxBlocks uxBlock.UxBlocks) int {
	if err == nil {
		return 0
	}
	uxBlocks.LogDebug(fmt.Sprintf("error: %+v", err))

	// A command that manages its own output and needs a specific exit status
	// (e.g. `upgrade --check`) signals it with an ExitError; return the code
	// as-is without printing anything further.
	if exitErr := errorsx.AsExitError(err); exitErr != nil {
		return exitErr.Code
	}

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintErrorText(err.Error())
		return 1
	}

	if apiErr, ok := errorsx.AsApiError(err); ok {
		uxBlocks.PrintErrorText(apiErr.Message())
		if apiErr.Meta() != nil {
			meta, err := yaml.Marshal(apiErr.Meta())
			if err != nil {
				uxBlocks.PrintErrorText(fmt.Sprintf("couldn't parse meta of error: %s", apiErr.Message()))
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

func createProfile(ctx context.Context, uxBlocks *uxBlock.Blocks, isTerminal bool) (*profile.Manager, error) {
	filePath, fileMode, err := constants.CliProfileFilePath()
	if err != nil {
		return nil, err
	}

	legacyPath, _, err := constants.CliLegacyDataFilePath()
	if err != nil {
		return nil, err
	}

	secretsPath := os.Getenv(constants.CliSecretsFilePathEnvVar)
	if secretsPath == "" {
		secretsPath = filepath.Join(filepath.Dir(filePath), constants.CliSecretsFileName)
	}

	envMode := secrets.Mode(os.Getenv(constants.SecretsBackendEnvVar))
	storedMode := peekStoredSecretsMode(filePath)

	mode := envMode
	if mode == "" {
		mode = storedMode
	}

	// Prompt on first run when nothing pins the backend and the user has a real choice. CI and
	// other non-TTY environments stay on the silent ModeAuto fallback so pipelines without a
	// keyring keep working without intervention.
	var persistChoice bool
	if envMode == "" && (storedMode == "" || storedMode == secrets.ModeAuto) && isTerminal && secrets.HasOSKeyring() {
		chosen, err := promptSecretsBackend(ctx, uxBlocks)
		if err != nil {
			return nil, err
		}
		mode = chosen
		persistChoice = true
	}

	secretsBackend, err := secrets.New(secrets.Config{
		Mode:     mode,
		Service:  "zerops-cli",
		FilePath: secretsPath,
		FileMode: 0o600,
		Log:      func(msg string) { uxBlocks.PrintWarningText(msg) },
	})
	if err != nil {
		return nil, err
	}

	manager, err := profile.NewManager(profile.Config{
		GlobalPath:       filePath,
		GlobalMode:       fileMode,
		LegacyGlobalPath: legacyPath,
		Cwd:              os.Getenv(constants.CliCwdEnvVar), // empty = Manager falls back to os.Getwd()
		Secrets:          secretsBackend,
		Log:              func(msg string) { uxBlocks.PrintInfoText(msg) },
	})
	if err != nil {
		return nil, err
	}

	if persistChoice {
		if _, err := manager.UpdateGlobal(func(d profile.Data) profile.Data {
			d.SecretsBackend = mode
			return d
		}); err != nil {
			return nil, errors.Wrap(err, "persist secrets backend preference")
		}
		if mode == secrets.ModeKeyring {
			// Stop here so the user reads the setup result instead of cobra's help / the
			// command's own output stacking on top. Re-running picks up the new backend
			// without going through any of this again.
			uxBlocks.PrintInfoText("Secrets storage configured: " + secretsBackend.Description() + ". Re-run your command to continue.")
			return nil, errorsx.NewExitError(0)
		}
	}

	return manager, nil
}

// promptSecretsBackend asks the user where sensitive values should live. Shown only on the
// first interactive run; the choice is persisted so subsequent runs skip the prompt.
func promptSecretsBackend(ctx context.Context, uxBlocks *uxBlock.Blocks) (secrets.Mode, error) {
	uxBlocks.PrintInfoText("zcli stores your auth token and VPN keys between runs. Choose where to keep them.")
	choices := []string{"OS keyring (recommended)", "Plaintext file"}
	descriptions := []string{
		"OS keyring: secrets are encrypted by your operating system (macOS Keychain, GNOME Keyring/KWallet, Windows Credential Manager) and unlocked with your login session. The OS may prompt for access the first time. Best on a normal desktop; needs a graphical session on Linux.",
		"Plaintext file: secrets are written to a chmod 0600 JSON file under the zcli config dir. No encryption at rest, but works without a desktop session and survives reboots silently. Use this for CI agents, headless servers, or if you prefer the file is easy to copy/inspect.",
	}
	idx, err := uxBlock.Run(
		prompt.NewRoot(ctx, "Where should zcli store secrets?", choices, prompt.WithChoiceDescriptions(descriptions)),
		prompt.GetChoiceCursor,
	)
	if err != nil {
		return "", err
	}
	if idx == 0 {
		uxBlocks.PrintInfoText("OS keyring selected.")
		uxBlocks.PrintInfoText("On the first secret zcli writes, your desktop's keyring service may ask you to create a new keyring and set a password for it. That password is owned by your operating system, not by zcli - zcli never sees it.")
		uxBlocks.PrintInfoText("The password is generally unrecoverable: losing it means the keyring's contents can't be decrypted and you may have to delete and recreate the keyring (and log in to zcli again). Pick something you'll remember, or that your password manager will.")
		if err := waitForEnter(ctx, "Press Enter to continue..."); err != nil {
			return "", err
		}
		return secrets.ModeKeyring, nil
	}
	return secrets.ModePlaintext, nil
}

// waitForEnterModel is a tiny Bubble Tea model that spins until the user presses Enter.
// Ctrl+C / Esc are reported as cancellation; tea.WithContext makes ctx cancellation (SIGINT
// or SIGTERM via the root signal handler) terminate the program too.
type waitForEnterModel struct {
	message string
	spinner bubblesSpinner.Model
	quit    bool
	err     error
}

func (m *waitForEnterModel) Init() tea.Cmd { return m.spinner.Tick }

func (m *waitForEnterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quit {
		return m, tea.Quit
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type { //nolint:exhaustive // only Enter / cancellation keys are meaningful here; everything else feeds the spinner.
		case tea.KeyEnter:
			m.quit = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = models.ErrCtrlC
			m.quit = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *waitForEnterModel) View() string {
	if m.quit {
		return ""
	}
	return m.spinner.View() + " " + m.message
}

func waitForEnter(ctx context.Context, message string) error {
	model := &waitForEnterModel{
		message: message,
		spinner: bubblesSpinner.New(bubblesSpinner.WithSpinner(bubblesSpinner.MiniDot)),
	}
	final, err := tea.NewProgram(model, tea.WithContext(ctx), tea.WithoutSignalHandler()).Run()
	if ctx.Err() != nil {
		return models.ErrCtrlC
	}
	if err != nil {
		return errors.Wrap(err, "wait for confirmation before keyring creation")
	}
	if fm, ok := final.(*waitForEnterModel); ok && fm.err != nil {
		return fm.err
	}
	return nil
}

// peekStoredSecretsMode reads only the secrets_backend field from the profile JSON, without
// triggering a full load or migration. Lets us resolve the user's preferred backend before
// constructing the secrets layer that the full load needs.
func peekStoredSecretsMode(profilePath string) secrets.Mode {
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		return secrets.ModeAuto
	}
	var probe struct {
		SecretsBackend secrets.Mode `json:"secrets_backend,omitempty"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || probe.SecretsBackend == "" {
		return secrets.ModeAuto
	}
	return probe.SecretsBackend
}
