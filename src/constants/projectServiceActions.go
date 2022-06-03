package constants

const (
	Success    = "✓ "
	Starting   = "→ "
	WorkingDir = "./"
)

type ParentCmd int

const (
	Project ParentCmd = iota
	Service
)
