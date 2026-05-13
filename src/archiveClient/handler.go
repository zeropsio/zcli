package archiveClient

const (
	WorkspaceAll    = "all"
	WorkspaceStaged = "staged"
	WorkspaceClean  = "clean"
)

type Config struct {
	Verbose            bool
	NoGit              bool // if set, the entire folder will be uploaded as is, without git commands
	DeployGitFolder    bool
	PushWorkspaceState string
}

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
	}
}
