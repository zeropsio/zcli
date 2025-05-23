package table

import (
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/gn"
)

type Body struct {
	rows []*Row
}

func NewBody(rows ...*Row) *Body {
	return &Body{
		rows: rows,
	}
}

func NewBodyFromStrings(rows ...[]string) *Body {
	return &Body{
		rows: gn.TransformSlice(rows, func(in []string) *Row {
			return NewRowFromStrings(in...)
		}),
	}
}

func (b *Body) Clone() *Body {
	clonedRows := make([]*Row, len(b.rows))
	for i, r := range b.rows {
		clonedRows[i] = r.Clone()
	}
	return &Body{rows: clonedRows}
}

func (b *Body) AddRow(row *Row) {
	if row.index < 0 {
		row.index = len(b.rows)
	}
	b.rows = append(b.rows, row)
}

func (b *Body) AddCellsRows(rows ...[]Cell) *Body {
	for _, cells := range rows {
		b.AddCellsRow(cells...)
	}
	return b
}

func (b *Body) AddStringsRows(rows ...[]string) *Body {
	for _, row := range rows {
		b.AddStringsRow(row...)
	}
	return b
}

func (b *Body) AddCellsRow(cells ...Cell) *Body {
	b.AddRow(NewRow(cells...))
	return b
}

func (b *Body) AddStringsRow(cells ...string) *Body {
	b.AddRow(NewRowFromStrings(cells...))
	return b
}

func (b *Body) Rows() []*Row {
	return b.rows
}

type Row struct {
	index int
	cells []Cell
	style lipgloss.Style
}

func NewRow(cells ...Cell) *Row {
	return &Row{
		index: -1,
		cells: cells,
	}
}

func NewRowFromStrings(cells ...string) *Row {
	return &Row{
		index: -1,
		cells: gn.TransformSlice(cells, stringsToCells),
	}
}

func stringsToCells(s string) Cell {
	return NewCell(s)
}

func (r *Row) Clone() *Row {
	c := *r
	c.cells = slices.Clone(r.cells)
	return &c
}

func (r *Row) AddCell(cell Cell) *Row {
	r.cells = append(r.cells, cell)
	return r
}

func (r *Row) AddCells(cells ...Cell) *Row {
	r.cells = append(r.cells, cells...)
	return r
}

func (r *Row) AddStringCells(cells ...string) *Row {
	for _, cell := range cells {
		r.AddStringCell(cell)
	}
	return r
}

func (r *Row) AddStringCell(text string) *Row {
	r.cells = append(r.cells, NewCell(text))
	return r
}

func (r *Row) Index() int {
	return r.index
}

func (r *Row) Cells() []Cell {
	return r.cells
}

type Cell struct {
	text  string
	style lipgloss.Style

	returnStyled bool
}

func NewCell(text string) Cell {
	return Cell{text: text}
}

func (c Cell) String() string {
	if c.returnStyled {
		return c.Styled()
	}
	return c.text
}

func (c Cell) SetPretty(styled bool) Cell {
	c.returnStyled = styled
	return c
}

func (c Cell) Styled() string {
	return c.style.Render(c.text)
}

func (c Cell) SetStyle(style lipgloss.Style) Cell {
	c.style = style
	return c
}
