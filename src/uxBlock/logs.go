package uxBlock

import (
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func (b *Blocks) GetOutputLogger() logger.Logger {
	return b.outputLogger
}

func (b *Blocks) GetDebugFileLogger() logger.Logger {
	return b.debugFileLogger
}

func (b *Blocks) LogDebug(message string) {
	b.debugFileLogger.Debug(message)
}

func (b *Blocks) PrintInfo(line styles.Line) {
	b.outputLogger.Info(line)
	b.debugFileLogger.Info(line.DisableStyle())
}

func (b *Blocks) PrintWarning(line styles.Line) {
	b.outputLogger.Warning(line)
	b.debugFileLogger.Warning(line.DisableStyle())
}

func (b *Blocks) PrintError(line styles.Line) {
	b.outputLogger.Error(line)
	b.debugFileLogger.Error(line.DisableStyle())
}
