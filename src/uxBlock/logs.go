package uxBlock

func (b *UxBlocks) PrintLine(values ...interface{}) {
	b.info(values...)
}

func (b *UxBlocks) PrintSuccessLine(values ...string) {
	b.info(SuccessIcon, successColor.SetString(values...))
}

func (b *UxBlocks) PrintInfoLine(values ...string) {
	b.info(InfoIcon, infoColor.SetString(values...))
}

func (b *UxBlocks) PrintWarningLine(values ...string) {
	b.warning(WarningIcon, warningColor.SetString(values...))
}

func (b *UxBlocks) PrintErrorLine(values ...string) {
	b.error(ErrorIcon, errorColor.SetString(values...))
}

func (b *UxBlocks) PrintDebugLine(args ...interface{}) {
	b.debugFileLogger.Debug(NewLine(args...).DisableStyle())
}

func (b *UxBlocks) info(args ...interface{}) {
	b.outputLogger.Info(NewLine(args...))
	b.debugFileLogger.Info(NewLine(args...).DisableStyle())
}

func (b *UxBlocks) warning(args ...interface{}) {
	b.outputLogger.Warning(NewLine(args...))
	b.debugFileLogger.Warning(NewLine(args...).DisableStyle())
}

func (b *UxBlocks) error(args ...interface{}) {
	b.outputLogger.Error(NewLine(args...))
	b.debugFileLogger.Error(NewLine(args...).DisableStyle())
}
