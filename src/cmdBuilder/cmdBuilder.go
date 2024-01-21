package cmdBuilder

type CmdBuilder struct {
	commands []*Cmd
}

func NewCmdBuilder() *CmdBuilder {
	return &CmdBuilder{}
}

func (b *CmdBuilder) AddCommand(cmd *Cmd) {
	b.commands = append(b.commands, cmd)
}
