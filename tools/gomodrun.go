package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

func getMod(dir string) (string, error) {
	if _, err := os.Stat(path.Join(dir, "go.mod")); err == nil {
		return dir, nil
	}
	newDir := path.Dir(dir)
	if newDir != "" {
		return getMod(newDir)
	}
	return "", fmt.Errorf("go.mod not found")
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	modDir, err := getMod(wd)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		log.Fatal("executable not defined")
		return
	}

	bin := os.Args[1]
	if !path.IsAbs(bin) {
		bin = path.Join(modDir, "bin", bin)
	}

	if err := os.Setenv("PATH", path.Join(modDir, "bin")+":"+os.Getenv("PATH")); err != nil {
		log.Fatal(err)
		return
	}

	cmd := &exec.Cmd{
		Path:   bin,
		Dir:    wd,
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		Args:   os.Args[1:],
		Env:    os.Environ(),
	}
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
