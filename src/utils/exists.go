package utils

import (
	"os"
)

func FileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil && !f.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
