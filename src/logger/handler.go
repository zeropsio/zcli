package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/zeropsio/zcli/src/i18n"

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

	l.AddHook(&StdoutHook{
		levels:    []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		formatter: formatter,
	})

	return &Handler{
		logrus: l,
	}
}

type DebugFileConfig struct {
	FilePath string
}

func NewDebugFileLogger(config DebugFileConfig) *Handler {
	l := logrus.New()
	l.Out = io.Discard
	l.Level = logrus.DebugLevel

	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err != nil {
		os.Stdout.WriteString(fmt.Sprintf(i18n.T(i18n.LoggerUnableToOpenLogFileWarning), config.FilePath))
	} else {
		file.Close()

		l.AddHook(&VarLogHook{
			path:   config.FilePath,
			levels: []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
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
