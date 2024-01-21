// Package uxBlock provides building blocks for UX and communication with a user.
package uxBlock

import (
	"context"

	"github.com/zeropsio/zcli/src/logger"
)

type UxBlocks struct {
	isTerminal      bool
	outputLogger    logger.Logger
	debugFileLogger logger.Logger

	// FIXME - janhajek comment
	ctxCancel context.CancelFunc
}

func NewBlock(
	outputLogger logger.Logger,
	debugFileLogger logger.Logger,
	isTerminal bool,
	ctxCancel context.CancelFunc,
) *UxBlocks {
	// safety check
	if ctxCancel == nil {
		ctxCancel = func() {}
	}

	return &UxBlocks{
		outputLogger:    outputLogger,
		debugFileLogger: debugFileLogger,
		isTerminal:      isTerminal,
		ctxCancel:       ctxCancel,
	}
}
