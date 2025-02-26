//go:build exclude

package archiveClient

import (
	"testing"

	"github.com/golang/mock/gomock"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/uxBlock/mocks"
)

var findByRulesTestCases = []struct {
	name        string
	workingDir  string
	input       []string
	output      []string
	ignoreLines []string
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
			"test/var/www/non–ascii.txt",
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
			"non–ascii.txt",
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

	// ignore file tests
	{
		name:       "ignore - no ignore",
		workingDir: "test2/var/www/",
		input: []string{
			"./",
		},
		output: []string{
			".env",
			".env.dist",
			"dir/",
			"dir/file4.ext1",
			"dir2/",
			"dir2/dev.env",
			"dir3/",
			"dir3/file3.ext2",
			"dir3/file4.ext2",
			"dir3/subDir/",
			"dir3/subDir/file5.ext1",
			"dir3/subDir/subSubDir/",
			"dir3/subDir/subSubDir/file6.ext2",
			"file1.ext1",
			"file2.ext2",
			"file3.ext2",
		},
	},
	{
		name:       "ignore - everything excluding file1.ext1",
		workingDir: "test2/var/www/",
		input: []string{
			"./",
		},
		ignoreLines: []string{
			"*",
			"!file1.ext1",
		},
		output: []string{
			"file1.ext1",
		},
	},
	{
		name:       "ignore - ext2 but not in dir3",
		workingDir: "test2/var/www/",
		input: []string{
			"./",
		},
		ignoreLines: []string{
			"*.ext2",
			"!dir3/*.ext2",
		},
		output: []string{
			".env",
			".env.dist",
			"dir/",
			"dir/file4.ext1",
			"dir2/",
			"dir2/dev.env",
			"dir3/",
			"dir3/file3.ext2",
			"dir3/file4.ext2",
			"dir3/subDir/",
			"dir3/subDir/file5.ext1",
			"dir3/subDir/subSubDir/",
			"file1.ext1",
		},
	},
	{
		name:       "ignore - dir wildcard",
		workingDir: "test2/var/www/",
		input: []string{
			"./",
		},
		ignoreLines: []string{
			"dir3/*/subSubDir",
		},
		output: []string{
			".env",
			".env.dist",
			"dir/",
			"dir/file4.ext1",
			"dir2/",
			"dir2/dev.env",
			"dir3/",
			"dir3/file3.ext2",
			"dir3/file4.ext2",
			"dir3/subDir/",
			"dir3/subDir/file5.ext1",
			"file1.ext1",
			"file2.ext2",
			"file3.ext2",
		},
	},
	{
		name:       "ignore - .env",
		workingDir: "test2/var/www/",
		input: []string{
			"./",
		},
		ignoreLines: []string{
			"*.env",
			"!.env.dist",
		},
		output: []string{
			".env.dist",
			"dir/",
			"dir/file4.ext1",
			"dir2/",
			"dir3/",
			"dir3/file3.ext2",
			"dir3/file4.ext2",
			"dir3/subDir/",
			"dir3/subDir/file5.ext1",
			"dir3/subDir/subSubDir/",
			"dir3/subDir/subSubDir/file6.ext2",
			"file1.ext1",
			"file2.ext2",
			"file3.ext2",
		},
	},
}

func TestFindFilesByRules(t *testing.T) {
	def, err := createNonAsciiFile()
	if err != nil {
		t.Fatal(err)
	}
	defer def()

	ctrl := gomock.NewController(t)
	uxBlocks := mocks.NewMockUxBlocks(ctrl)
	uxBlocks.EXPECT().PrintInfo(gomock.Any()).AnyTimes()

	for _, test := range findByRulesTestCases {
		test := test // scope lint
		t.Run(test.name+"-in-"+test.workingDir, func(t *testing.T) {
			archiver := New(Config{})

			var ignorer FileIgnorer
			if test.ignoreLines != nil {
				ignorer = ignore.CompileIgnoreLines(test.ignoreLines...)
			}

			files, err := archiver.FindFilesByRules(uxBlocks, test.workingDir, test.input, ignorer)
			require.NoError(t, err)

			output := make([]string, 0, len(files))
			for _, f := range files {
				output = append(output, f.ArchivePath)
			}

			require.Equal(t, test.output, output)
		})
	}
}
