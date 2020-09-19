package utils

import (
	"bufio"
	"io"
	"os"
)

func SetFirstLine(path string, firstLine string) error {
	lines, err := ReadLines(path)
	if err != nil {
		return err
	}
	if len(lines) == 0 || lines[0] != firstLine {
		lines = append([]string{firstLine}, lines...)
		err := WriteLines(path, lines)
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveFirstLine(path string, firstLine string) error {
	lines, err := ReadLines(path)
	if err != nil {
		return err
	}
	if len(lines) > 0 && lines[0] == firstLine {
		lines = lines[1:]
		err := WriteLines(path, lines)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadLines(path string) (lines []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rd := bufio.NewReader(f)

	for {
		lineB, _, err := rd.ReadLine()
		line := string(lineB)

		if err == io.EOF {
			if line != "" {
				lines = append(lines, line)
			}
			break
		}
		if err != nil {
			return nil, err
		}

		lines = append(lines, line)
	}

	return
}

func WriteLines(path string, lines []string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, line := range lines {
		if _, err = f.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}
