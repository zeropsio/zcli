package archiveClient

import (
	"os"
	"path"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

type FileIgnorer interface {
	MatchesPath(string) bool
}

const DeployIgnoreFile = ".deployignore"

// LoadDeployFileIgnorer parses .deployignore file in specified dir.
// If file is absent, both returned FileIgnorer and error is nil.
func LoadDeployFileIgnorer(dir string) (FileIgnorer, error) {
	absFilepath, err := filepath.Abs(path.Join(dir, DeployIgnoreFile))
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(absFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return ignore.CompileIgnoreFile(absFilepath)
}
