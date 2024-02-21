package uxBlock

import "github.com/zeropsio/zcli/src/uxBlock/styles"

func (b *uxBlocks) LogDebug(message string) {
	b.debugFileLogger.Debug(message)
}

func (b *uxBlocks) PrintInfo(line styles.Line) {
	b.outputLogger.Info(line)
	b.debugFileLogger.Info(line.DisableStyle())
}

func (b *uxBlocks) PrintWarning(line styles.Line) {
	b.outputLogger.Warning(line)
	b.debugFileLogger.Warning(line.DisableStyle())
}

func (b *uxBlocks) PrintError(line styles.Line) {
	b.outputLogger.Error(line)
	b.debugFileLogger.Error(line.DisableStyle())
}
