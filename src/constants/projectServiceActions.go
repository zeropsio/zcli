package constants

const (
	Success  = "✓ "
	Starting = "→ "
)

type ParentCmd int

const (
	Project ParentCmd = iota
	Service
)
