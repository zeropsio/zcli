package constants

const (
	Start    = "start"
	Stop     = "stop"
	Delete   = "delete"
	Success  = "✓ "
	Starting = "→ "
)

type ParentCmd int

const (
	Service ParentCmd = iota
	Project
)
