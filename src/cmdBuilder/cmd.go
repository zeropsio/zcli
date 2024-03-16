package cmdBuilder

import (
	"context"
)

type loggedUserRunFunc func(ctx context.Context, cmdData *LoggedUserCmdData) error
type guestRunFunc func(ctx context.Context, cmdData *GuestCmdData) error

type ScopeLevel interface {
	AddCommandFlags(*Cmd)
	LoadSelectedScope(ctx context.Context, cmd *Cmd, cmdData *LoggedUserCmdData) error
	GetParent() ScopeLevel
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

func (cmd *Cmd) AddChildrenCmd(childrenCmd *Cmd) *Cmd {
	cmd.childrenCmds = append(cmd.childrenCmds, childrenCmd)
	return cmd
}

func (cmd *Cmd) Use(use string) *Cmd {
	cmd.use = use
	return cmd
}

func (cmd *Cmd) SetHelpTemplate(template string) *Cmd {
	cmd.helpTemplate = template
	return cmd
}

func (cmd *Cmd) Short(short string) *Cmd {
	cmd.short = short
	return cmd
}

func (cmd *Cmd) Long(long string) *Cmd {
	cmd.long = long
	return cmd
}

func (cmd *Cmd) LoggedUserRunFunc(runFunc loggedUserRunFunc) *Cmd {
	cmd.loggedUserRunFunc = runFunc
	return cmd
}

func (cmd *Cmd) GuestRunFunc(runFunc guestRunFunc) *Cmd {
	cmd.guestRunFunc = runFunc
	return cmd
}

func (cmd *Cmd) SilenceUsage(silenceUsage bool) *Cmd {
	cmd.silenceUsage = silenceUsage
	return cmd
}

func (cmd *Cmd) SilenceError(silenceError bool) *Cmd {
	cmd.silenceError = silenceError
	return cmd
}

func (cmd *Cmd) ScopeLevel(scopeLevel ScopeLevel) *Cmd {
	cmd.scopeLevel = scopeLevel
	return cmd
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

func (cmd *Cmd) Arg(name string, auxOptions ...ArgOption) *Cmd {
	cfg := cmdArg{
		name: name,
	}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	cmd.args = append(cmd.args, cfg)
	return cmd
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

func (cmd *Cmd) StringFlag(name string, defaultValue string, description string, auxOptions ...FlagOption) *Cmd {
	return cmd.addFlag(name, defaultValue, description, auxOptions...)
}

func (cmd *Cmd) IntFlag(name string, defaultValue int, description string, auxOptions ...FlagOption) *Cmd {
	return cmd.addFlag(name, defaultValue, description, auxOptions...)
}

func (cmd *Cmd) BoolFlag(name string, defaultValue bool, description string, auxOptions ...FlagOption) *Cmd {
	return cmd.addFlag(name, defaultValue, description, auxOptions...)
}

func (cmd *Cmd) HelpFlag(description string, auxOptions ...FlagOption) *Cmd {
	auxOptions = append(auxOptions, ShortHand("h"))
	return cmd.addFlag("help", false, description, auxOptions...)
}

func (cmd *Cmd) addFlag(name string, defaultValue interface{}, description string, auxOptions ...FlagOption) *Cmd {
	cfg := cmdFlag{
		name:         name,
		description:  description,
		defaultValue: defaultValue,
	}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	cmd.flags = append(cmd.flags, cfg)
	return cmd
}
