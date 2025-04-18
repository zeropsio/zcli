package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/zeropsio/zcli/src/file"
)

type VarLogHook struct {
	path      string
	fileMode  os.FileMode
	levels    []logrus.Level
	formatter logrus.Formatter

	lock sync.Mutex
}

func (hook *VarLogHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *VarLogHook) Fire(entry *logrus.Entry) error {
	hook.lock.Lock()
	defer hook.lock.Unlock()

	msg, err := hook.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	f, err := file.Open(hook.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, hook.fileMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open file, %v", err)
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(msg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write message, %v", err)
		return err
	}

	return nil
}

type TerminalHook struct {
	levels     []logrus.Level
	formatter  logrus.Formatter
	isTerminal bool
}

func (hook *TerminalHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *TerminalHook) Fire(entry *logrus.Entry) error {
	msg := []byte(entry.Message)
	if hook.formatter != nil {
		if formattedEntry, err := hook.formatter.Format(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		} else {
			msg = formattedEntry
		}
	} else {
		msg = append(msg, '\n')
	}

	if entry.Level <= logrus.ErrorLevel {
		os.Stderr.Write(msg)
	} else {
		os.Stdout.Write(msg)
	}
	return nil
}
