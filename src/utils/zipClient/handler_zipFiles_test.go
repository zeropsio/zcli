package zipClient

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSymlink(t *testing.T) {
	RegisterTestingT(t)

	ziper := New(Config{})

	b := &bytes.Buffer{}
	err := ziper.ZipFiles(
		b,
		[]File{
			{
				SourcePath:  "./test/var/www/dir/subDir/file3.3.symlink.txt",
				ArchivePath: "dir/subDir/file3.3.symlink.txt",
			},
		},
	)
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
