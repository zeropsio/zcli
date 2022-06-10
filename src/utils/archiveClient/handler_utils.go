package archiveClient

import (
	"os"
	"path/filepath"
	"strings"
)

// fixes paths/dirs that may be missing between source root and deployed files (all dirs are needed for valid TAR file)
// provided createFile func will receive a valid File.SourcePath and MUST strip workDir to form a valid File.ArchivePath
// createdPath is used as a cache to not create paths multiple times for multiple calls of the function
func (h Handler) fixMissingDirPath(files []File, createFile func(filePath string) File, createdPaths map[string]struct{}) []File {
	fixedFiles := make([]File, 0, len(files)+50)

	for _, file := range files {
		// filepath.Dir calls Clean, which replaces all / with os.PathSeparator (same as calling filepath.FromSlash)
		dirPath := filepath.Dir(file.ArchivePath)
		if dirPath == "." {
			fixedFiles = append(fixedFiles, file)
			continue
		}

		dirPath += string(os.PathSeparator)
		if dirPath == file.ArchivePath {
			createdPaths[dirPath] = struct{}{}
			fixedFiles = append(fixedFiles, file)
			continue
		}

		if _, ok := createdPaths[dirPath]; ok {
			fixedFiles = append(fixedFiles, file)
			continue
		}

		// path must start with valid source path prefix (createFile must strip it from File.ArchivePath if necessary)
		path := strings.TrimSuffix(file.SourcePath, filepath.FromSlash(file.ArchivePath))

		// split path into separate folders, to correctly create every folder
		paths := strings.Split(dirPath, string(os.PathSeparator))
		for _, p := range paths {
			path = filepath.Join(path, p) + string(os.PathSeparator)
			if _, ok := createdPaths[path]; ok {
				continue
			}
			missingFile := createFile(path)
			if missingFile.ArchivePath == "" {
				continue
			}
			fixedFiles = append(fixedFiles, missingFile)
			createdPaths[path] = struct{}{}
		}

		fixedFiles = append(fixedFiles, file)
	}

	return fixedFiles
}
