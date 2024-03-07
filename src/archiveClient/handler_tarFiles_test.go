package archiveClient

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSymlink(t *testing.T) {
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
	require.NoError(t, err)

	b, err := io.ReadAll(gz)
	require.NoError(t, err)

	r := tar.NewReader(bytes.NewReader(b))
	require.NoError(t, err)

	for {
		header, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		switch header.Typeflag {
		case tar.TypeSymlink:
			require.Equal(t, "../file2.1.txt", header.Linkname)
		default:
			t.Fatal("unknown type")
		}
	}
}
