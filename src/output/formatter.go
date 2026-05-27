package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeropsio/zcli/src/uxBlock/models/table"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "", "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	default:
		return FormatTable, fmt.Errorf("unsupported output format %q (supported: %s)", s, SupportedFormats())
	}
}

func SupportedFormats() string {
	return "table, json, csv"
}

func PrintData(headers []string, rows [][]string, format Format) (string, error) {
	switch format {
	case FormatJSON:
		return printJSON(headers, rows)
	case FormatCSV:
		return printCSV(headers, rows)
	default:
		return printTable(headers, rows), nil
	}
}

func printTable(headers []string, rows [][]string) string {
	headerRow := table.NewRowFromStrings(headers...)
	body := table.NewBody()
	for _, row := range rows {
		body.AddStringsRow(row...)
	}
	return table.Render(body, table.WithHeader(headerRow))
}

func printJSON(headers []string, rows [][]string) (string, error) {
	items := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		item := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				item[h] = row[i]
			}
		}
		items = append(items, item)
	}
	out, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func printCSV(headers []string, rows [][]string) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	if err := w.Write(headers); err != nil {
		return "", err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	return b.String(), w.Error()
}
