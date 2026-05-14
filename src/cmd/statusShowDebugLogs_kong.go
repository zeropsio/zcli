package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// ShowDebugLogsCmd tails the last ~10 lines of the debug log file. Guest run.
type ShowDebugLogsCmd struct{}

func (*ShowDebugLogsCmd) Run(_ context.Context, d *kongCmd.CmdData) error {
	logFilePath, fileMode, err := constants.LogFilePath()
	if err != nil {
		return err
	}

	f, err := file.Open(logFilePath, os.O_RDONLY, fileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	filesize := stat.Size()
	if filesize == 0 {
		d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.DebugLogsNotFound)))
		return nil
	}

	var (
		line   string
		cursor int64
		lines  []string
	)
	for {
		cursor--
		if _, err := f.Seek(cursor, io.SeekEnd); err != nil {
			return errors.WithStack(err)
		}

		char := make([]byte, 1)
		if _, err := f.Read(char); err != nil {
			return errors.WithStack(err)
		}

		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			if len(lines) > 10 {
				break
			}
			lines = append([]string{line}, lines...)
			line = ""
		}

		line = fmt.Sprintf("%s%s", string(char), line)

		if cursor == -filesize {
			lines = append([]string{line}, lines...)
			break
		}
	}

	for _, line := range lines {
		d.Stderr.Print(line)
	}
	return nil
}
