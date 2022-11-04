package vpn

import (
	"sync"
	"time"

	"github.com/zeropsio/zcli/src/daemonStorage"
	"github.com/zeropsio/zcli/src/utils/logger"
)

type Config struct {
	VpnCheckInterval   time.Duration
	VpnCheckRetryCount int
	VpnCheckTimeout    time.Duration
}

type Handler struct {
	config  Config
	logger  logger.Logger
	storage *daemonStorage.Handler

	lock sync.RWMutex
}

func New(
	config Config,
	logger logger.Logger,
	daemonStorage *daemonStorage.Handler,

) *Handler {
	return &Handler{
		config:  config,
		logger:  logger,
		storage: daemonStorage,
	}
}
