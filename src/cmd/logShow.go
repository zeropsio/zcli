package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func logShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show",
		Short:        i18n.CmdLogShow,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			f, err := os.OpenFile(constants.LogFilePath, os.O_RDONLY, 0777)
			if err != nil {
				return err
			}

			line := ""
			var cursor int64 = 0
			stat, _ := f.Stat()
			filesize := stat.Size()

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

				line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

				if cursor == -filesize { // stop if we are at the begining
					lines = append([]string{line}, lines...)
					break
				}
			}

			for _, line := range lines {
				fmt.Print(line)
			}

			return nil
		},
	}

	return cmd
}
