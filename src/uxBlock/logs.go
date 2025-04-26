package uxBlock

import (
	"bufio"
	"bytes"

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

func (b *Blocks) PrintInfoText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintInfo(styles.InfoLine(scan.Text()))
	}
}

func (b *Blocks) PrintInfo(line styles.Line) {
	b.outputLogger.Info(line)
	b.debugFileLogger.Info(line.DisableStyle())
}

func (b *Blocks) PrintWarningText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintWarning(styles.WarningLine(scan.Text()))
	}
}

func (b *Blocks) PrintWarning(line styles.Line) {
	b.outputLogger.Warning(line)
	b.debugFileLogger.Warning(line.DisableStyle())
}

func (b *Blocks) PrintErrorText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintError(styles.ErrorLine(scan.Text()))
	}
}

func (b *Blocks) PrintError(line styles.Line) {
	b.outputLogger.Error(line)
	b.debugFileLogger.Error(line.DisableStyle())
}
