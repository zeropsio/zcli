package selector

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Option = generic.Option[Model]

func WithLabel(label string) Option {
	return func(m *Model) {
		m.label = optional.New(styles.SelectLine(label).String())
	}
}

func WithSetEnableMultiSelect(b bool) Option {
	return func(m *Model) {
		m.multi = b
		m.keyMap.MultiSelect.SetEnabled(b)
	}
}

func WithHeader(header *table.Row) Option {
	return func(m *Model) {
		m.header = header
	}
}

func WithSetEnableFiltering(b bool) Option {
	return func(m *Model) {
		m.enableFiltering = b
		m.keyMap.Filter.SetEnabled(b)
	}
}

func New(tableBody *table.Body, opts ...Option) *Model {
	return generic.ApplyOptionsWithDefault(
		Model{
			keyMap:      DefaultKeymap(),
			tableBody:   tableBody,
			selected:    make(map[int]struct{}),
			jump:        5,
			filterField: textinput.New(),
		},
		opts...,
	)
}

type Model struct {
	keyMap    KeyMap
	header    *table.Row
	tableBody *table.Body
	cursor    int
	selected  map[int]struct{}

	enableFiltering bool
	filterField     textinput.Model
	filteredBody    *table.Body

	label optional.Null[string]
	multi bool
	jump  int

	width  int
	height int
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) IsMultiSelect() bool {
	return m.multi
}

func (m *Model) Selected() []int {
	selection := generic.TransformMapToSlice(m.selected, func(k int, v struct{}) int {
		return k
	})
	if len(selection) == 1 {
		return selection
	}
	slices.Sort(selection)
	return selection
}

func (m *Model) Init() tea.Cmd {
	m.filterField.Prompt = ""
	m.filterField.Placeholder = "press / to filter"
	m.filterField.PlaceholderStyle = styles.HelpStyle()
	m.filteredBody = m.tableBody.Clone()
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.filterField.Focused() {
		m.filter()
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if key.Matches(keyMsg, m.keyMap.Select) {
				m.filterField.Blur()
				m.cursor = 0
			}
		}
		m.filterField, cmd = m.filterField.Update(msg)
		return m, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keyMap.LineUp):
			m.Up(1)
		case key.Matches(keyMsg, m.keyMap.LineDown):
			m.Down(1)
		case key.Matches(keyMsg, m.keyMap.PageUp):
			m.Up(m.jump)
		case key.Matches(keyMsg, m.keyMap.PageDown):
			m.Down(m.jump)
		case key.Matches(keyMsg, m.keyMap.Select):
			rowIndex := m.filteredBody.Rows()[m.cursor].Index()
			m.selected[rowIndex] = struct{}{}
			return m, Selected
		case key.Matches(keyMsg, m.keyMap.MultiSelect):
			rowIndex := m.filteredBody.Rows()[m.cursor].Index()
			if _, ok := m.selected[rowIndex]; ok {
				delete(m.selected, rowIndex)
			} else {
				m.selected[rowIndex] = struct{}{}
			}
		case key.Matches(keyMsg, m.keyMap.Filter):
			return m, m.filterField.Focus()
		}
	}

	return m, nil
}

func (m *Model) Up(jump int) {
	m.cursor = max(0, m.cursor-jump)
}

func (m *Model) Down(jump int) {
	m.cursor = min(len(m.filteredBody.Rows())-1, m.cursor+jump)
}

func (m *Model) filter() {
	if m.filterField.Value() == "" {
		m.filteredBody = m.tableBody.Clone()
		return

	}
	f := table.NewBody()
	for _, r := range m.tableBody.Rows() {
		_, hasMatchingCell := generic.FindOne(r.Cells(), func(in table.Cell) bool {
			cell := in.SetPretty(false).String()
			return strings.Contains(strings.ToLower(cell), m.filterField.Value())
		})
		if hasMatchingCell {
			f.AddRow(r.Clone())
		}
	}
	m.filteredBody = f
}

func (m *Model) View() string {
	rows := m.filteredBody.Rows()
	totalRows := len(rows)

	// FIXME lhellmann: take care with this comment
	// we have 1 row with filter, so always -1, and we also need to remove 2 more rows for top and bottom border
	maxRows := m.height - 5
	// if label "Please select xyz" is present -1
	if m.label.Filled() {
		maxRows -= 1
	}
	// if header is present, it takes additional 2 rows (1 for text and 1 for border between header and content)
	if m.header != nil {
		maxRows -= 2
	}
	// remove 1 more row for pagination
	if totalRows > maxRows {
		maxRows -= 1
	}
	if maxRows <= 0 {
		return "Your terminal window is too small to render the table."
	}

	t := ltable.New().
		Border(lipgloss.RoundedBorder()).
		BorderColumn(false).
		BorderRow(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == m.cursor%maxRows && !m.filterField.Focused() {
				return styles.TableRowActive()
			}

			return styles.TableRow()
		})

	var columns int
	if m.header != nil {
		headers := generic.TransformSlice(m.header.Cells(), func(in table.Cell) string {
			return strings.ToUpper(in.String())
		})
		t = t.Headers(headers...)
		columns = len(headers)
	}

	m.appendRows(t, rows, totalRows, maxRows, columns)
	// addPager := func(currentPage, maxPages int) {
	// 	t = t.Row(fmt.Sprintf("%d/%d", currentPage, maxPages))
	// }
	//
	// maxPages := int(math.Ceil(float64(totalRows) / float64(maxRows)))
	// for rowIndex, row := range rows {
	// 	page := m.cursor / maxRows // starts at 0 for math operations (needs +1 for rendering)
	//
	// 	// do not render rows at the beginning of the table if we go past them
	// 	if rowIndex < page*maxRows {
	// 		continue
	// 	}
	// 	// do not render rows past the limit
	// 	if rowIndex >= (page+1)*maxRows {
	// 		addPager(page+1, maxPages)
	// 		break
	// 	}
	//
	// 	cells := generic.TransformSlice(
	// 		row.Cells(),
	// 		func(in table.Cell) string {
	// 			return in.String()
	// 		})
	//
	// 	fillCells(columns, cells, "-")
	//
	// 	t = t.Row(cells...)
	//
	// 	// Add pager to the last page
	// 	if page > 0 && rowIndex == len(rows)-1 {
	// 		addPager(page+1, maxPages)
	// 	}
	// }

	t.Width(calculateTableWidth(t, m.width))

	var s string
	if label, set := m.label.Get(); set {
		s = label
	}
	if m.enableFiltering {
		s = lipgloss.JoinVertical(lipgloss.Left, s,
			m.filterField.View(),
		)
	}
	if s != "" {
		return lipgloss.JoinVertical(lipgloss.Left,
			s,
			t.String(),
		)
	}
	return t.String()
}

func (m *Model) appendRows(t *ltable.Table, rows []*table.Row, totalRows, maxRows, columns int) {
	addPager := func(currentPage, maxPages int) {
		t.Row(fmt.Sprintf("%d/%d", currentPage, maxPages))
	}

	maxPages := int(math.Ceil(float64(totalRows) / float64(maxRows)))
	for rowIndex, row := range rows {
		page := m.cursor / maxRows // starts at 0 for math operations (needs +1 for rendering)

		// do not render rows at the beginning of the table if we go past them
		if rowIndex < page*maxRows {
			continue
		}
		// do not render rows past the limit
		if rowIndex >= (page+1)*maxRows {
			addPager(page+1, maxPages)
			break
		}

		cells := generic.TransformSlice(
			row.Cells(),
			func(in table.Cell) string {
				return in.String()
			})

		fillCells(columns, cells, "-")

		t = t.Row(cells...)

		// Add pager to the last page
		if page > 0 && rowIndex == len(rows)-1 {
			addPager(page+1, maxPages)
		}
	}
}

func fillCells(columns int, cells []string, filler string) {
	numOfCells := len(cells)
	if columns > numOfCells {
		for i := 0; i < columns-numOfCells; i++ {
			cells = append(cells, filler)
		}
	}
}

// calculateTableWidth calculates the width of the table.
// If the table is wider than the terminal, a table starts falling apart.
// To set a fix width could help, but in that case, even if the table is smaller, it takes the whole terminal width.
// And it doesn't look good.
func calculateTableWidth(t *ltable.Table, terminalWidth int) int {
	tableWidth := lipgloss.Width(t.String())
	if tableWidth > terminalWidth {
		return terminalWidth
	}
	return 0
}
