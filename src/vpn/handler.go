package vpn

import (
	"sync"
	"time"

	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
	"github.com/zerops-io/zcli/src/utils/logger"
)

type Config struct {
	VpnCheckInterval   time.Duration
	VpnCheckRetryCount int
	VpnCheckTimeout    time.Duration
}

type Handler struct {
	config    Config
	logger    logger.Logger
	storage   *daemonStorage.Handler
	dnsServer *dnsServer.Handler

	lock sync.RWMutex
}

func New(
	config Config,
	logger logger.Logger,
	daemonStorage *daemonStorage.Handler,
	dnsServer *dnsServer.Handler,

) *Handler {
	return &Handler{
		config:    config,
		logger:    logger,
		storage:   daemonStorage,
		dnsServer: dnsServer,
	}
}
