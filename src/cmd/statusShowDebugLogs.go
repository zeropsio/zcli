package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
)

func statusShowDebugLogsCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("show-debug-logs").
		Short(i18n.T(i18n.CmdStatusShowDebugLogs)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			logFilePath, err := constants.LogFilePath()
			if err != nil {
				return err
			}

			f, err := os.OpenFile(logFilePath, os.O_RDONLY, 0777)
			if err != nil {
				return err
			}

			line := ""
			var cursor int64 = 0
			stat, _ := f.Stat()
			filesize := stat.Size()

			if filesize == 0 {
				// FIXME - janhajek translate
				fmt.Println("No logs found")
				return nil
			}

			lines := []string{}
			for {
				cursor -= 1
				f.Seek(cursor, io.SeekEnd)

				char := make([]byte, 1)
				f.Read(char)

				if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
					if len(lines) > 10 {
						break
					}
					lines = append([]string{line}, lines...)
					line = ""
				}

				line = fmt.Sprintf("%s%s", string(char), line)

				if cursor == -filesize { // stop if we are at the begining
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
