package uxBlock

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
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

func (b *uxBlocks) Table(body *TableBody, auxOptions ...TableOption) {
	cfg := tableConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	t := table.New().
		BorderStyle(styles.TableBorderStyle()).
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return styles.TableRow()
		})

	if cfg.header != nil {
		headers := make([]string, len(cfg.header.cells))
		for i, header := range cfg.header.cells {
			headers[i] = strings.ToUpper(header.Text)
		}
		t = t.Headers(headers...)
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

	t.Width(calculateTableWidth(t, b.terminalWidth))

	fmt.Println(t)
}

// calculateTableWidth calculates the width of the table.
// If the table is wider than the terminal, a table starts falling apart.
// To set a fix width could help, but in that case, even if the table is smaller, it takes the whole terminal width.
// And it doesn't look good.
func calculateTableWidth(t *table.Table, terminalWidth int) int {
	tableWidth := lipgloss.Width(t.String())
	if tableWidth > terminalWidth {
		return terminalWidth
	}
	return 0
}
