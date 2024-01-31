package uxBlock

func (b *uxBlocks) PrintLine(values ...interface{}) {
	b.info(values...)
}

func (b *uxBlocks) PrintSuccessLine(values ...string) {
	b.info(SuccessIcon, successColor.SetString(values...))
}

func (b *uxBlocks) PrintInfoLine(values ...string) {
	b.info(InfoIcon, infoColor.SetString(values...))
}

func (b *uxBlocks) PrintWarningLine(values ...string) {
	b.warning(WarningIcon, warningColor.SetString(values...))
}

func (b *uxBlocks) PrintErrorLine(values ...string) {
	b.error(ErrorIcon, errorColor.SetString(values...))
}

func (b *uxBlocks) PrintDebugLine(args ...interface{}) {
	b.debugFileLogger.Debug(NewLine(args...).DisableStyle())
}

func (b *uxBlocks) info(args ...interface{}) {
	b.outputLogger.Info(NewLine(args...))
	b.debugFileLogger.Info(NewLine(args...).DisableStyle())
}

func (b *uxBlocks) warning(args ...interface{}) {
	b.outputLogger.Warning(NewLine(args...))
	b.debugFileLogger.Warning(NewLine(args...).DisableStyle())
}

func (b *uxBlocks) error(args ...interface{}) {
	b.outputLogger.Error(NewLine(args...))
	b.debugFileLogger.Error(NewLine(args...).DisableStyle())
}
