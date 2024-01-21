package uxBlock

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type tableConfig struct {
	header *TableRow
}

type TableBody struct {
	rows []*TableRow
}

func NewTableBody() *TableBody {
	return &TableBody{}
}

func (b *TableBody) AddRow(row *TableRow) {
	b.rows = append(b.rows, row)
}

func (b *TableBody) AddStringsRows(rows ...[]string) *TableBody {
	for _, row := range rows {
		b.AddStringsRow(row...)
	}

	return b
}

func (b *TableBody) AddStringsRow(cells ...string) *TableBody {
	b.AddRow(NewTableRow().AddStringCells(cells...))

	return b
}

type TableRow struct {
	cells []*TableCell
}

func NewTableRow() *TableRow {
	return &TableRow{}
}

func (r *TableRow) AddCell(cell *TableCell) *TableRow {
	r.cells = append(r.cells, cell)

	return r
}

func (r *TableRow) AddStringCells(cells ...string) *TableRow {
	for _, cell := range cells {
		r.AddStringCell(cell)
	}

	return r
}

func (r *TableRow) AddStringCell(text string) *TableRow {
	r.cells = append(r.cells, NewTableCell(text))

	return r
}

type TableCell struct {
	Text string
}

func NewTableCell(text string) *TableCell {
	return &TableCell{Text: text}
}

func WithTableHeader(header *TableRow) TableOption {
	return func(cfg *tableConfig) {
		cfg.header = header
	}
}

type TableOption = func(cfg *tableConfig)

func (b *UxBlocks) Table(body *TableBody, auxOptions ...TableOption) {
	cfg := tableConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Copy().Foreground(lipgloss.Color("252")).Bold(true)

	baseStyle.SetString()

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 && cfg.header != nil {
				return headerStyle
			}

			even := row%2 == 0

			if even {
				return baseStyle.Copy().Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Copy().Foreground(lipgloss.Color("252"))
		})

	if cfg.header != nil {
		capitalizeHeaders := func(data []string) []string {
			for i := range data {
				data[i] = strings.ToUpper(data[i])
			}
			return data
		}

		headers := make([]string, len(cfg.header.cells))
		for i, header := range cfg.header.cells {
			headers[i] = header.Text
		}
		t = t.Headers(capitalizeHeaders(headers)...)
	}

	rows := make([][]string, len(body.rows))
	for rowIndex, row := range body.rows {
		cells := make([]string, len(row.cells))
		for i, cell := range row.cells {
			cells[i] = cell.Text
		}
		rows[rowIndex] = cells
	}
	t = t.Rows(rows...)

	fmt.Println(t)
}
