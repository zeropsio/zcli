package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	logrus *logrus.Logger
}

type OutputConfig struct {
	IsTerminal bool
}

func NewOutputLogger(config OutputConfig) *Handler {
	l := logrus.New()
	l.Out = io.Discard
	l.Level = logrus.DebugLevel

	var formatter logrus.Formatter
	if !config.IsTerminal {
		formatter = &logrus.TextFormatter{DisableColors: true}
	}

	l.AddHook(&TerminalHook{
		levels:    []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		formatter: formatter,
	})

	return &Handler{
		logrus: l,
	}
}

type DebugFileConfig struct {
	FilePath string
	FileMode os.FileMode
}

func NewDebugFileLogger(config DebugFileConfig) *Handler {
	l := logrus.New()
	l.Out = io.Discard
	l.Level = logrus.DebugLevel

	if config.FilePath != "" {
		l.AddHook(&VarLogHook{
			path:     config.FilePath,
			fileMode: config.FileMode,
			levels:   []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
			formatter: &logrus.TextFormatter{
				DisableColors: true,
			},
		})
	}

	return &Handler{
		logrus: l,
	}
}

func (h *Handler) Info(a ...interface{}) {
	h.logrus.Info(a...)
}

func (h *Handler) Infof(format string, a ...interface{}) {
	h.logrus.Infof(format, a...)
}

func (h *Handler) Warning(a ...interface{}) {
	h.logrus.Warning(a...)
}

func (h *Handler) Warningf(format string, a ...interface{}) {
	h.logrus.Warningf(format, a...)
}

func (h *Handler) Error(a ...interface{}) {
	h.logrus.Error(a...)
}

func (h *Handler) Errorf(format string, a ...interface{}) {
	h.logrus.Errorf(format, a...)
}

func (h *Handler) Debug(a ...interface{}) {
	h.logrus.Debug(a...)
}

func (h *Handler) Debugf(format string, a ...interface{}) {
	h.logrus.Debugf(format, a...)
}

func (h *Handler) Write(p []byte) (n int, err error) {
	return h.logrus.Writer().Write(p)
}
