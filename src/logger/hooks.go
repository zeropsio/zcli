package logger

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type VarLogHook struct {
	path      string
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

	f, err := os.OpenFile(hook.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open file, %v", err)
		return err
	}

	defer f.Close()

	_, err = io.WriteString(f, string(msg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write message, %v", err)
		return err
	}

	return nil
}

type StdoutHook struct {
	levels    []logrus.Level
	formatter logrus.Formatter
}

func (hook *StdoutHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *StdoutHook) Fire(entry *logrus.Entry) error {
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
