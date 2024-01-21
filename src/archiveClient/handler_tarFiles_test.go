package archiveClient

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSymlink(t *testing.T) {
	RegisterTestingT(t)

	archiver := New(Config{})
	errChan := make(chan error)
	reader, writer := io.Pipe()

	go archiver.TarFiles(
		writer,
		[]File{
			{
				SourcePath:  "./test/var/www/dir/subDir/file3.3.symlink.txt",
				ArchivePath: "dir/subDir/file3.3.symlink.txt",
			},
		},
		errChan,
	)

	gz, err := gzip.NewReader(reader)
	Expect(err).ShouldNot(HaveOccurred())

	b, err := io.ReadAll(gz)
	Expect(err).ShouldNot(HaveOccurred())

	r := tar.NewReader(bytes.NewReader(b))
	Expect(err).ShouldNot(HaveOccurred())

	for {
		header, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		Expect(err).ShouldNot(HaveOccurred())

		switch header.Typeflag {
		case tar.TypeSymlink:
			Expect(header.Linkname).To(Equal("../file2.1.txt"))
		default:
			Expect(errors.New("unknown type")).ShouldNot(HaveOccurred())
		}
	}
}
