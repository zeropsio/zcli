package constants

const (
	Success  = "✓ "
	Starting = "→ "
)

type ChildCmd int

const (
	Start ChildCmd = iota
	Stop
	Delete
)

type ParentCmd int

const (
	Service ParentCmd = iota
	Project
)
