package test

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/zerops-io/zcli/src/service/zipClient"

	. "github.com/onsi/gomega"
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
			"./",
		},
		output: []string{
			"var/",
			"var/www/",
			"var/www/dir/",
			"var/www/dir/file2.1.txt",
			"var/www/dir/file2.2.txt",
			"var/www/dir/subDir/",
			"var/www/dir/subDir/file3.1.txt",
			"var/www/dir/subDir/file3.2.txt",
			"var/www/dir/subDir/file3.3.symlink.txt",
			"var/www/file1.1.txt",
			"zip_test.go",
		},
	},
	{
		name:       "single files",
		workingDir: "./",
		input: []string{
			"var/www/file1.1.txt",
			"var/www/dir/file2.1.txt",
			"var/www/dir/subDir/file3.1.txt",
		},
		output: []string{
			"var/www/file1.1.txt",
			"var/www/dir/file2.1.txt",
			"var/www/dir/subDir/file3.1.txt",
		},
	},
	{
		name:       "all files in directory",
		workingDir: "./",
		input: []string{
			"var/www/dir",
		},
		output: []string{
			"var/www/dir/",
			"var/www/dir/file2.1.txt",
			"var/www/dir/file2.2.txt",
			"var/www/dir/subDir/",
			"var/www/dir/subDir/file3.1.txt",
			"var/www/dir/subDir/file3.2.txt",
			"var/www/dir/subDir/file3.3.symlink.txt",
		},
	},
	{
		name:       "all files in sub directory",
		workingDir: "./",
		input: []string{
			"var/www/dir/subDir",
		},
		output: []string{
			"var/www/dir/subDir/",
			"var/www/dir/subDir/file3.1.txt",
			"var/www/dir/subDir/file3.2.txt",
			"var/www/dir/subDir/file3.3.symlink.txt",
		},
	},
	{
		name:       "single files - strip directory",
		workingDir: "./",
		input: []string{
			"var/www/dir/*/file2.1.txt",
			"var/www/dir/*/subDir/file3.1.txt",
			"var/www/dir/subDir/*/file3.1.txt",
		},
		output: []string{
			"file2.1.txt",
			"subDir/file3.1.txt",
			"file3.1.txt",
		},
	},
	{
		name:       "all files - strip directory",
		workingDir: "./",
		input: []string{
			"var/www/dir/*",
			"var/www/dir/subDir/*",
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

	//////////////////////
	// with working dir
	/////////////////////

	{
		name:       "all",
		workingDir: "var/www/",
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
		workingDir: "var/www/",
		input: []string{
			"file1.1.txt",
			"dir/file2.1.txt",
			"dir/subDir/file3.1.txt",
		},
		output: []string{
			"file1.1.txt",
			"dir/file2.1.txt",
			"dir/subDir/file3.1.txt",
		},
	},
	{
		name:       "all files in directory",
		workingDir: "var/www/",
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
		workingDir: "var/www/",
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
		workingDir: "var/www/",
		input: []string{
			"dir/*/file2.1.txt",
			"dir/*/subDir/file3.1.txt",
			"dir/subDir/*/file3.1.txt",
		},
		output: []string{
			"file2.1.txt",
			"subDir/file3.1.txt",
			"file3.1.txt",
		},
	},
	{
		name:       "all files - strip directory",
		workingDir: "var/www/",
		input: []string{
			"dir/*",
			"dir/subDir/*",
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
	for _, test := range testErrorResponseDataProvider {
		test := test // scope lint
		t.Run(test.name+" in "+test.workingDir, func(t *testing.T) {
			RegisterTestingT(t)

			logger := debugLogger{}

			ziper := zipClient.New(zipClient.Config{}, logger)

			b := &bytes.Buffer{}
			err := ziper.Zip(b, test.workingDir, test.input...)
			Expect(err).ShouldNot(HaveOccurred())

			r, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(len(b.Bytes())))
			Expect(err).ShouldNot(HaveOccurred())

			output := func() (res []string) {
				for _, f := range r.File {
					res = append(res, f.Name)
				}
				return
			}()

			Expect(output).To(Equal(test.output))
		})
	}
}

func TestSymlink(t *testing.T) {
	RegisterTestingT(t)

	logger := debugLogger{}

	ziper := zipClient.New(zipClient.Config{}, &logger)

	b := &bytes.Buffer{}
	err := ziper.Zip(b, "var/www/", "dir/subDir/file3.3.symlink.txt")
	Expect(err).ShouldNot(HaveOccurred())

	r, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(len(b.Bytes())))
	Expect(err).ShouldNot(HaveOccurred())

	Expect(r.File).To(HaveLen(1))

	fo, err := r.File[0].Open()
	Expect(err).ShouldNot(HaveOccurred())
	defer fo.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, fo)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(buf.String()).To(Equal("../file2.1.txt"))

}

type debugLogger struct {
}

func (d debugLogger) Info(a ...interface{}) {

}

func (d debugLogger) Warning(a ...interface{}) {

}

func (d debugLogger) Error(a ...interface{}) {

}

func (d debugLogger) Debug(a ...interface{}) {

}
