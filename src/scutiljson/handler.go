package scutiljson

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errEndOfBlock = errors.New("eob")

func parseDict(scanner *bufio.Scanner) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		key, value, err := parseLine(scanner, line)
		if err != nil {
			if err != errEndOfBlock {
				return nil, err
			}
			return res, nil
		}
		res[key] = value
	}
	return res, nil
}

func parseArray(scanner *bufio.Scanner) ([]interface{}, error) {
	res := make([]interface{}, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		_, value, err := parseLine(scanner, line)
		if err != nil {
			if err != errEndOfBlock {
				return nil, err
			}
			return res, nil
		}
		res = append(res, value)
	}

	return res, nil
}

func parseLine(scanner *bufio.Scanner, line string) (key string, value interface{}, err error) {
	parts := strings.SplitN(line, " : ", 2)
	if len(parts) == 1 {
		if line[len(line)-1] == '}' {
			return "", nil, errEndOfBlock
		}
		return "", nil, fmt.Errorf("do not know how to parse: %s", line)
	}

	switch p := parts[1]; p {
	case "No such key":
		return "", nil, nil
	case "<dictionary> {":
		value, err = parseDict(scanner)
	case "<array> {":
		value, err = parseArray(scanner)
	case "true":
		value = true
	default:
		value = p
	}

	return strings.TrimSpace(parts[0]), value, err
}

func Encode(r io.Reader) (map[string]interface{}, error) {
	br := bufio.NewReader(r)
	firstLine, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	firstLine = strings.TrimRight(firstLine, "\n\r ")
	topKey := strings.TrimSuffix(firstLine, " <dictionary> {")
	if topKey == firstLine {
		scanner := bufio.NewScanner(br)
		return parseDict(scanner)
	}
	topMap := make(map[string]interface{})
	scanner := bufio.NewScanner(br)
	res, err := parseDict(scanner)
	if err != nil {
		return nil, err
	}
	topMap[topKey] = res
	return topMap, nil
}

// JSONEncode reads from r which contains a scutil --nc formatted data
// and writes the equivalent JSON structure to w.
func JSONEncode(r io.Reader, w io.Writer) error {
	topMap, err := Encode(r)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&topMap); err != nil {
		return fmt.Errorf("error encoding json: %s", err)
	}
	return nil
}
