package archiveClient

import (
	"github.com/zeropsio/zcli/src/logger"
)

const (
	WorkspaceAll    = "all"
	WorkspaceStaged = "staged"
	WorkspaceClean  = "clean"
)

type Config struct {
	Logger             logger.Logger
	Verbose            bool
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
