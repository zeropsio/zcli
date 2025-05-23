package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Option = gn.Option[tableConfig]
type tableConfig struct {
	header        *Row
	width, height int
}

func WithHeader(header *Row) Option {
	return func(cfg *tableConfig) {
		cfg.header = header
	}
}

func WithSize(width, height int) Option {
	return func(cfg *tableConfig) {
		cfg.width, cfg.height = width, height
	}
}

func Render(body *Body, opts ...Option) string {
	cfg := gn.ApplyOptions(opts...)

	t := table.New().
		BorderStyle(styles.TableBorderStyle()).
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return styles.TableRow()
		})

	if cfg.header != nil {
		headers := make([]string, len(cfg.header.Cells()))
		for i, header := range cfg.header.Cells() {
			headers[i] = strings.ToUpper(header.String())
		}
		t = t.Headers(headers...)
	}

	rows := make([][]string, len(body.rows))
	for rowIndex, row := range body.rows {
		cells := make([]string, len(row.Cells()))
		for i, cell := range row.Cells() {
			cells[i] = cell.String()
		}
		rows[rowIndex] = cells
	}
	t = t.Rows(rows...)

	t.Width(min(lipgloss.Width(t.String()), cfg.width))

	return t.String()
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
