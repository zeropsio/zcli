package archiveClient

const (
	WorkspaceAll    = "all"
	WorkspaceStaged = "staged"
	WorkspaceClean  = "clean"
)

type Config struct {
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
