package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func statusShowDebugLogsCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("show-debug-logs").
		Short(i18n.T(i18n.CmdStatusShowDebugLogs)).
		HelpFlag(i18n.T(i18n.StatusShowDebugLogsHelp)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			logFilePath, fileMode, err := constants.LogFilePath()
			if err != nil {
				return err
			}

			f, err := file.Open(logFilePath, os.O_RDONLY, fileMode)
			if err != nil {
				return err
			}
			defer f.Close()

			line := ""
			var cursor int64 = 0
			stat, _ := f.Stat()
			filesize := stat.Size()

			if filesize == 0 {
				cmdData.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.DebugLogsNotFound)))
				return nil
			}

			lines := []string{}
			for {
				cursor -= 1
				_, err = f.Seek(cursor, io.SeekEnd)
				if err != nil {
					return errors.WithStack(err)
				}

				char := make([]byte, 1)
				_, err = f.Read(char)
				if err != nil {
					return errors.WithStack(err)
				}

				if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
					if len(lines) > 10 {
						break
					}
					lines = append([]string{line}, lines...)
					line = ""
				}

				line = fmt.Sprintf("%s%s", string(char), line)

				if cursor == -filesize { // stop if we are at the beginning
					lines = append([]string{line}, lines...)
					break
				}
			}

			for _, line := range lines {
				fmt.Print(line)
			}

			return nil
		})
}
