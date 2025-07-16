package cmdBuilder

import (
	"bufio"
	"context"
	"strings"
	"unicode/utf8"

	"github.com/zeropsio/zcli/src/terminal"
)

type loggedUserRunFunc func(ctx context.Context, cmdData *LoggedUserCmdData) error
type guestRunFunc func(ctx context.Context, cmdData *GuestCmdData) error

type ScopeLevel interface {
	AddCommandFlags(*Cmd)
	LoadSelectedScope(ctx context.Context, cmd *Cmd, cmdData *LoggedUserCmdData) error
}

type Cmd struct {
	use               string
	short             string
	long              string
	helpTemplate      string
	loggedUserRunFunc loggedUserRunFunc
	guestRunFunc      guestRunFunc
	silenceUsage      bool
	silenceError      bool

	scopeLevel ScopeLevel
	args       []cmdArg
	flags      []cmdFlag

	childrenCmds []*Cmd
}

type cmdArg struct {
	name          string
	optional      bool
	isArray       bool
	optionalLabel string
}

type cmdFlag struct {
	name         string
	defaultValue interface{}
	description  string
	hidden       bool
	shorthand    string
}

func NewCmd() *Cmd {
	return &Cmd{
		silenceUsage: true,
	}
}

func (c *Cmd) AddChildrenCmd(childrenCmd *Cmd) *Cmd {
	c.childrenCmds = append(c.childrenCmds, childrenCmd)
	return c
}

func (c *Cmd) Use(use string) *Cmd {
	c.use = use
	return c
}

func (c *Cmd) SetHelpTemplate(template string) *Cmd {
	c.helpTemplate = template
	return c
}

func (c *Cmd) Short(short string) *Cmd {
	c.short = short
	return c
}

func (c *Cmd) Long(long string) *Cmd {
	c.long = long
	return c
}

func (c *Cmd) LoggedUserRunFunc(runFunc loggedUserRunFunc) *Cmd {
	c.loggedUserRunFunc = runFunc
	return c
}

func (c *Cmd) GuestRunFunc(runFunc guestRunFunc) *Cmd {
	c.guestRunFunc = runFunc
	return c
}

func (c *Cmd) SilenceUsage(silenceUsage bool) *Cmd {
	c.silenceUsage = silenceUsage
	return c
}

func (c *Cmd) SilenceError(silenceError bool) *Cmd {
	c.silenceError = silenceError
	return c
}

func (c *Cmd) ScopeLevel(scopeLevel ScopeLevel) *Cmd {
	c.scopeLevel = scopeLevel
	return c
}

type ArgOption = func(cfg *cmdArg)

func OptionalArg() ArgOption {
	return func(cfg *cmdArg) {
		cfg.optional = true
	}
}

func ArrayArg() ArgOption {
	return func(cfg *cmdArg) {
		cfg.isArray = true
	}
}

func OptionalArgLabel(label string) ArgOption {
	return func(cfg *cmdArg) {
		cfg.optionalLabel = label
	}
}

func (c *Cmd) Arg(name string, auxOptions ...ArgOption) *Cmd {
	cfg := cmdArg{
		name: name,
	}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	c.args = append(c.args, cfg)
	return c
}

type FlagOption = func(cfg *cmdFlag)

func HiddenFlag() FlagOption {
	return func(cfg *cmdFlag) {
		cfg.hidden = true
	}
}

func ShortHand(shorthand string) FlagOption {
	return func(cfg *cmdFlag) {
		cfg.shorthand = shorthand
	}
}

func (c *Cmd) StringFlag(name string, defaultValue string, description string, auxOptions ...FlagOption) *Cmd {
	return c.addFlag(name, defaultValue, description, auxOptions...)
}

func (c *Cmd) StringSliceFlag(name string, defaultValue []string, description string, auxOptions ...FlagOption) *Cmd {
	return c.addFlag(name, defaultValue, description, auxOptions...)
}

func (c *Cmd) IntFlag(name string, defaultValue int, description string, auxOptions ...FlagOption) *Cmd {
	return c.addFlag(name, defaultValue, description, auxOptions...)
}

func (c *Cmd) BoolFlag(name string, defaultValue bool, description string, auxOptions ...FlagOption) *Cmd {
	return c.addFlag(name, defaultValue, description, auxOptions...)
}

func (c *Cmd) HelpFlag(description string, auxOptions ...FlagOption) *Cmd {
	auxOptions = append(auxOptions, ShortHand("h"))
	return c.addFlag("help", false, description, auxOptions...)
}

func (c *Cmd) addFlag(name string, defaultValue interface{}, description string, auxOptions ...FlagOption) *Cmd {
	cfg := cmdFlag{
		name:         name,
		description:  warpStringForTerminal(description),
		defaultValue: defaultValue,
	}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	c.flags = append(c.flags, cfg)
	return c
}

func warpStringForTerminal(s string) string {
	width, _ := terminal.GetSize()
	if width < 0 {
		return s
	}
	maxRunes := min(70, uint(float64(width)*.30))
	return warpString(s, maxRunes)
}

// warpString wraps string after num of runes (excluding space as rune), keeping words intact
func warpString(s string, runeCutCount uint) string {
	if s == "" {
		return ""
	}
	output := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)
	var runeCount uint
	for scanner.Scan() {
		//nolint:gosec
		runeCount += uint(utf8.RuneCount(scanner.Bytes()))
		if runeCount >= runeCutCount {
			runeCount = 0
			output.WriteString(scanner.Text())
			output.WriteRune('\n')
			continue
		}
		output.WriteString(scanner.Text())
		output.WriteRune(' ')
	}
	return output.String()[:output.Len()-1]
}
