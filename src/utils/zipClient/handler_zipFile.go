package zipClient

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

func zipFile(archive *zip.Writer, file File, info os.FileInfo) error {

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	archivePath := file.ArchivePath

	if info.IsDir() {
		archivePath = strings.TrimSuffix(archivePath, string(os.PathSeparator)) + string(os.PathSeparator)
	}

	archivePath = strings.TrimPrefix(archivePath, string(os.PathSeparator))

	header.Name = archivePath

	if !info.IsDir() {
		header.Method = zip.Deflate
	}

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	if header.FileInfo().Mode()&os.ModeSymlink != 0 {
		symlink, err := os.Readlink(file.SourcePath)
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(symlink))
		if err != nil {
			return err
		}
	} else {
		file, err := os.Open(file.SourcePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}

	return nil
}
