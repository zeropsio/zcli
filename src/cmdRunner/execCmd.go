package cmdRunner

import (
	"context"
	"os/exec"
)

type Func func(ctx context.Context) error

func CommandContext(ctx context.Context, cmd string, args ...string) *ExecCmd {
	return &ExecCmd{
		Cmd: exec.CommandContext(ctx, cmd, args...),
		ctx: ctx,
	}
}

type ExecCmd struct {
	*exec.Cmd
	ctx    context.Context
	before Func
	after  Func
}

func (e *ExecCmd) SetBefore(f Func) *ExecCmd {
	e.before = f
	return e
}

func (e *ExecCmd) execBefore() error {
	if e.before == nil {
		return nil
	}
	return e.before(e.ctx)
}

func (e *ExecCmd) SetAfter(f Func) *ExecCmd {
	e.after = f
	return e
}

func (e *ExecCmd) execAfter() error {
	if e.after == nil {
		return nil
	}
	return e.after(e.ctx)
}
