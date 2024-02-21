package archiveClient

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/uxBlock/mocks"
)

var testErrorResponseDataProvider = []struct {
	name       string
	workingDir string
	input      []string
	output     []string
}{

	{
		name:       "all",
		workingDir: "./",
		input: []string{
			"./test/",
		},
		output: []string{
			"test/",
			"test/var/",
			"test/var/www/",
			"test/var/www/dir/",
			"test/var/www/dir/file2.1.txt",
			"test/var/www/dir/file2.2.txt",
			"test/var/www/dir/subDir/",
			"test/var/www/dir/subDir/file3.1.txt",
			"test/var/www/dir/subDir/file3.2.txt",
			"test/var/www/dir/subDir/file3.3.symlink.txt",
			"test/var/www/file1.1.txt",
		},
	},
	{
		name:       "single files",
		workingDir: "./",
		input: []string{
			"test/var/www/file1.1.txt",
			"test/var/www/dir/file2.1.txt",
			"test/var/www/dir/subDir/file3.1.txt",
		},
		output: []string{
			"test/",
			"test/var/",
			"test/var/www/",
			"test/var/www/file1.1.txt",
			"test/var/www/dir/",
			"test/var/www/dir/file2.1.txt",
			"test/var/www/dir/subDir/",
			"test/var/www/dir/subDir/file3.1.txt",
		},
	},
	{
		name:       "all files in directory",
		workingDir: "./",
		input: []string{
			"test/var/www/dir",
		},
		output: []string{
			"test/var/www/dir/",
			"test/var/www/dir/file2.1.txt",
			"test/var/www/dir/file2.2.txt",
			"test/var/www/dir/subDir/",
			"test/var/www/dir/subDir/file3.1.txt",
			"test/var/www/dir/subDir/file3.2.txt",
			"test/var/www/dir/subDir/file3.3.symlink.txt",
		},
	},
	{
		name:       "all files in sub directory",
		workingDir: "./",
		input: []string{
			"test/var/www/dir/subDir",
		},
		output: []string{
			"test/var/www/dir/subDir/",
			"test/var/www/dir/subDir/file3.1.txt",
			"test/var/www/dir/subDir/file3.2.txt",
			"test/var/www/dir/subDir/file3.3.symlink.txt",
		},
	},
	{
		name:       "single files - strip directory",
		workingDir: "./",
		input: []string{
			"test/var/www/dir/~/file2.1.txt",
			"test/var/www/dir/~/subDir/file3.1.txt",
			"test/var/www/dir/subDir/~/file3.1.txt",
		},
		output: []string{
			"file2.1.txt",
			"subDir/",
			"subDir/file3.1.txt",
			"file3.1.txt",
		},
	},
	{
		name:       "all files - strip directory",
		workingDir: "./",
		input: []string{
			"test/var/www/dir/~",
			"test/var/www/dir/subDir/~",
		},
		output: []string{
			"file2.1.txt",
			"file2.2.txt",
			"subDir/",
			"subDir/file3.1.txt",
			"subDir/file3.2.txt",
			"subDir/file3.3.symlink.txt",
			"file3.1.txt",
			"file3.2.txt",
			"file3.3.symlink.txt",
		},
	},

	// ////////////////////
	// with working dir
	// ///////////////////

	{
		name:       "all",
		workingDir: "test/var/www/",
		input: []string{
			"./",
		},
		output: []string{
			"dir/",
			"dir/file2.1.txt",
			"dir/file2.2.txt",
			"dir/subDir/",
			"dir/subDir/file3.1.txt",
			"dir/subDir/file3.2.txt",
			"dir/subDir/file3.3.symlink.txt",
			"file1.1.txt",
		},
	},
	{
		name:       "single files",
		workingDir: "test/var/www/",
		input: []string{
			"file1.1.txt",
			"dir/file2.1.txt",
			"dir/subDir/file3.1.txt",
		},
		output: []string{
			"file1.1.txt",
			"dir/",
			"dir/file2.1.txt",
			"dir/subDir/",
			"dir/subDir/file3.1.txt",
		},
	},
	{
		name:       "all files in directory",
		workingDir: "test/var/www/",
		input: []string{
			"dir",
		},
		output: []string{
			"dir/",
			"dir/file2.1.txt",
			"dir/file2.2.txt",
			"dir/subDir/",
			"dir/subDir/file3.1.txt",
			"dir/subDir/file3.2.txt",
			"dir/subDir/file3.3.symlink.txt",
		},
	},
	{
		name:       "all files in sub directory",
		workingDir: "test/var/www/",
		input: []string{
			"dir/subDir",
		},
		output: []string{
			"dir/subDir/",
			"dir/subDir/file3.1.txt",
			"dir/subDir/file3.2.txt",
			"dir/subDir/file3.3.symlink.txt",
		},
	},
	{
		name:       "single files - strip directory",
		workingDir: "test/var/www/",
		input: []string{
			"dir/~/file2.1.txt",
			"dir/~/subDir/file3.1.txt",
			"dir/subDir/~/file3.1.txt",
		},
		output: []string{
			"file2.1.txt",
			"subDir/",
			"subDir/file3.1.txt",
			"file3.1.txt",
		},
	},
	{
		name:       "all files - strip directory",
		workingDir: "test/var/www/",
		input: []string{
			"dir/~",
			"dir/subDir/~",
		},
		output: []string{
			"file2.1.txt",
			"file2.2.txt",
			"subDir/",
			"subDir/file3.1.txt",
			"subDir/file3.2.txt",
			"subDir/file3.3.symlink.txt",
			"file3.1.txt",
			"file3.2.txt",
			"file3.3.symlink.txt",
		},
	},
}

func TestValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	uxBlocks := mocks.NewMockUxBlocks(ctrl)
	uxBlocks.EXPECT().PrintInfo(gomock.Any()).AnyTimes()

	for _, test := range testErrorResponseDataProvider {
		test := test // scope lint
		t.Run(test.name+" in "+test.workingDir, func(t *testing.T) {
			archiver := New(Config{})

			files, err := archiver.FindFilesByRules(uxBlocks, test.workingDir, test.input)
			require.NoError(t, err)

			output := func() (res []string) {
				for _, f := range files {
					res = append(res, f.ArchivePath)
				}
				return
			}()

			require.Equal(t, test.output, output)
		})
	}
}
