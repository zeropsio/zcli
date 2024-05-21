package archiveClient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var findGitFilesTestCases = []struct {
	name       string
	workingDir string
	output     []string
}{
	{
		name:       "non-ascii",
		workingDir: "test/var/www",
		output: []string{
			"dir/",
			"dir/file2.1.txt",
			"dir/file2.2.txt",
			"dir/subDir/",
			"dir/subDir/file3.1.txt",
			"dir/subDir/file3.2.txt",
			"dir/subDir/file3.3.symlink.txt",
			"file1.1.txt",
			"non–ascii.txt",
		},
	},
}

func TestFindGitFiles(t *testing.T) {
	ctx := context.TODO()
	for _, test := range findGitFilesTestCases {
		t.Run(test.name+"-in-"+test.workingDir, func(t *testing.T) {
			assert := require.New(t)
			archiver := New(Config{})

			files, err := archiver.FindGitFiles(ctx, test.workingDir)
			assert.NoError(err)

			output := make([]string, 0, len(files))
			for _, f := range files {
				output = append(output, f.ArchivePath)
			}

			assert.Equal(test.output, output)
		})
	}
}
