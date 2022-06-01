package archiveClient

type Config struct {
	DeployGitFolder bool
}

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
	}
}
