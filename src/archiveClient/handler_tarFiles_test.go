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

	buf := bytes.NewBuffer(nil)

	require.NoError(t, archiver.TarFiles(
		buf,
		[]File{
			{
				SourcePath:  "./test/var/www/dir/subDir/file3.3.symlink.txt",
				ArchivePath: "dir/subDir/file3.3.symlink.txt",
			},
		},
	))

	gz, err := gzip.NewReader(buf)
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

func TestTarFiles_WithProgress(t *testing.T) {
	// WithProgress must fire once per file, monotonically, with the
	// running processed count and the constant total. Order matches
	// the input slice.
	archiver := New(Config{})
	files := []File{
		{SourcePath: "./test/var/www/file1.1.txt", ArchivePath: "file1.1.txt"},
		{SourcePath: "./test/var/www/dir/file2.1.txt", ArchivePath: "dir/file2.1.txt"},
		{SourcePath: "./test/var/www/dir/subDir/file3.1.txt", ArchivePath: "dir/subDir/file3.1.txt"},
	}

	type call struct{ processed, total int }
	var calls []call
	require.NoError(t, archiver.TarFiles(
		bytes.NewBuffer(nil),
		files,
		WithProgress(func(processed, total int) {
			calls = append(calls, call{processed, total})
		}),
	))

	require.Equal(t, []call{
		{1, 3}, {2, 3}, {3, 3},
	}, calls, "callback fires once per file with running count + constant total")
}

func TestTarFiles_WithoutProgressIsNoOp(t *testing.T) {
	// Backwards-compat: no opts means no callback machinery — the
	// existing tests (TestSymlink) already cover this, but pin it
	// explicitly to discourage accidentally making the callback
	// mandatory later.
	archiver := New(Config{})
	require.NoError(t, archiver.TarFiles(
		bytes.NewBuffer(nil),
		[]File{{SourcePath: "./test/var/www/file1.1.txt", ArchivePath: "file1.1.txt"}},
	))
}
