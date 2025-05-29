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
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Option = gn.Option[Model]

func WithLabel(label string) Option {
	return func(m *Model) {
		m.label = optional.New(styles.SelectLine(label).String())
	}
}

func WithSetEnableMultiSelect(enable bool) Option {
	return func(m *Model) {
		m.multi = enable
		m.keyMap.MultiSelect.SetEnabled(enable)
		m.keyMap.SelectAll.SetEnabled(enable)
		m.keyMap.DeselectAll.SetEnabled(enable)
	}
}

func WithEnableMultiSelect() Option {
	return func(m *Model) {
		WithSetEnableMultiSelect(true)(m)
	}
}

func WithHeader(header *table.Row) Option {
	return func(m *Model) {
		m.header = header
	}
}

func WithSetEnableFiltering(enable bool) Option {
	return func(m *Model) {
		m.enableFiltering = enable
		m.keyMap.Filter.SetEnabled(enable)
		m.keyMap.FilterClear.SetEnabled(enable)
		m.filterFunc = ExactFilter
	}
}

func WithEnableFiltering() Option {
	return func(m *Model) {
		WithSetEnableFiltering(true)(m)
	}
}

func WithFilterFunc(filterFunc FilterFunc) Option {
	return func(m *Model) {
		m.filterFunc = filterFunc
		if filterFunc == nil {
			m.filterFunc = filterNone
		}
	}
}

type FilterFunc func(string, string) bool

func filterNone(string, string) bool {
	return true
}

var ExactFilter = FilterFunc(strings.Contains)

type Model struct {
	keyMap    KeyMap
	header    *table.Row
	tableBody *table.Body
	cursor    int
	selected  map[int]struct{}

	enableFiltering bool
	filterField     textinput.Model
	filteredBody    *table.Body
	filterFunc      FilterFunc

	label optional.Null[string]
	multi bool
	jump  int

	width  int
	height int
}

func New(tableBody *table.Body, opts ...Option) *Model {
	return gn.ApplyOptionsWithDefault(
		Model{
			keyMap:      DefaultKeymap(),
			tableBody:   tableBody,
			selected:    make(map[int]struct{}),
			jump:        5,
			filterField: textinput.New(),
			filterFunc:  filterNone, // default for safety
		},
		opts...,
	)
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) IsMultiSelect() bool {
	return m.multi
}

func (m *Model) Selected() []int {
	selection := gn.TransformMapToSlice(m.selected, func(k int, v struct{}) int {
		return k
	})
	if len(selection) == 1 {
		return selection
	}
	slices.Sort(selection)
	return selection
}

func (m *Model) Init() tea.Cmd {
	m.filterField.CharLimit = 64
	m.filterField.Prompt = ""
	m.filterField.Placeholder = "press / to filter"
	m.filterField.PlaceholderStyle = styles.HelpStyle()
	m.filteredBody = m.tableBody.Clone()
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd tea.Cmd
	)

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
		case key.Matches(keyMsg, m.keyMap.Home):
			m.Home()
		case key.Matches(keyMsg, m.keyMap.End):
			m.End()
		case key.Matches(keyMsg, m.keyMap.Select):
			if !m.multi {
				m.Select()
			}
			return m, m.SelectCommand
		case key.Matches(keyMsg, m.keyMap.MultiSelect):
			m.Select()
		case key.Matches(keyMsg, m.keyMap.SelectAll, m.keyMap.DeselectAll):
			m.SelectAll(key.Matches(keyMsg, m.keyMap.DeselectAll))
		case key.Matches(keyMsg, m.keyMap.Filter):
			return m, m.filterField.Focus()
		case key.Matches(keyMsg, m.keyMap.FilterClear):
			m.filterField.SetValue("")
			m.filter()
			m.filterField.Blur()
		}
	}

	return m, nil
}

func (m *Model) Up(jump int) {
	if !m.hasResults() {
		return
	}
	m.cursor = max(0, m.cursor-jump)
}

func (m *Model) Down(jump int) {
	if !m.hasResults() {
		return
	}
	m.cursor = min(len(m.filteredBody.Rows())-1, m.cursor+jump)
}

func (m *Model) Home() {
	if !m.hasResults() {
		return
	}
	m.cursor = 0
}

func (m *Model) End() {
	if !m.hasResults() {
		return
	}
	m.cursor = len(m.filteredBody.Rows()) - 1
}

func (m *Model) Select() {
	if !m.hasResults() {
		return
	}
	selectedRow := m.filteredBody.Rows()[m.cursor]
	if m.multi {
		if _, ok := m.selected[selectedRow.Index()]; ok {
			delete(m.selected, selectedRow.Index())
			return
		}
		m.selected[selectedRow.Index()] = struct{}{}
		return
	}
	m.selected[selectedRow.Index()] = struct{}{}
}

func (m *Model) SelectAll(deselect bool) {
	for _, r := range m.filteredBody.Rows() {
		if !deselect {
			m.selected[r.Index()] = struct{}{}
		} else {
			delete(m.selected, r.Index())
		}
	}
}

func (m *Model) hasResults() bool {
	return len(m.filteredBody.Rows()) != 0
}

func (m *Model) filter() {
	if m.filterField.Value() == "" {
		m.filteredBody = m.tableBody.Clone()
		return
	}
	f := table.NewBody()
	for _, r := range m.tableBody.Rows() {
		_, hasMatchingCell := gn.FindFirst(r.Cells(), func(in table.Cell) bool {
			cell := in.SetPretty(false).String()
			return m.filterFunc(strings.ToLower(cell), m.filterField.Value())
		})
		if hasMatchingCell {
			f.AddRow(r.Clone())
		}
	}
	m.filteredBody = f
}

func (m *Model) isSelected(row int) bool {
	_, selected := m.selected[row]
	return selected
}

func (m *Model) View() string {
	rows := m.filteredBody.Rows()
	totalRows := len(rows)

	// -1 for filter line, and -2 for top and bottom border
	maxRows := m.height - 3
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
			if row == ltable.HeaderRow {
				return styles.TableRow()
			}
			if row == m.cursor%maxRows && !m.filterField.Focused() {
				return styles.TableRowActive()
			}
			return styles.TableRow()
		})

	var columns int
	if m.header != nil {
		columns = m.makeHeader(t)
	}

	if totalRows != 0 {
		m.makeRows(t, rows, totalRows, maxRows, columns)
	} else {
		m.emptyRows(t, columns)
	}

	t.Width(calculateTableWidth(t, m.width))

	var s string
	if label, set := m.label.Get(); set {
		s = label
	}
	if m.enableFiltering && s != "" {
		s = lipgloss.JoinVertical(lipgloss.Left,
			s,
			m.filterField.View(),
		)
	} else {
		s = m.filterField.View()
	}
	if s != "" {
		return lipgloss.JoinVertical(lipgloss.Left,
			s,
			t.String(),
		)
	}
	return t.String()
}

// symbols from https://symbl.cc/en/unicode-table/#geometric-shapes
const (
	checkMark = "✓" // https://symbl.cc/en/2713/
	filler    = ""

	unselected = "○" // https://symbl.cc/en/25CB/
	selected   = "◉" // https://symbl.cc/en/25C9/
)

func (m *Model) makeRows(t *ltable.Table, rows []*table.Row, totalRows, maxRows, columns int) {
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

		cells := gn.TransformSlice(
			row.Cells(),
			func(in table.Cell) string {
				return in.String()
			})

		cells = fillCells(columns, cells, filler)

		if m.multi {
			if totalRows > 0 && !m.isSelected(rows[rowIndex].Index()) {
				cells = gn.Prepend(cells, unselected)
			} else {
				cells = gn.Prepend(cells, selected)
			}
		}
		t.Row(cells...)

		// Add pager to the last page
		if page > 0 && rowIndex == len(rows)-1 {
			addPager(page+1, maxPages)
		}
	}
}

func (m *Model) makeHeader(t *ltable.Table) int {
	if !m.hasResults() {
		t.Headers("x")
		return 1
	}
	header := gn.TransformSlice(m.header.Cells(), func(in table.Cell) string {
		return strings.ToUpper(in.String())
	})
	numOfCells := len(header)
	if m.multi {
		header = gn.Prepend(header, checkMark)
	}
	t.Headers(header...)
	return numOfCells
}

func (m *Model) emptyRows(t *ltable.Table, columns int) {
	row := []string{styles.WarningText("No matches found").String()}
	if m.multi {
		columns++
	}
	row = fillCells(columns, row, filler)
	t.Row(row...)
}

//nolint:makezero
func fillCells(columns int, cells []string, filler string) []string {
	numOfCells := len(cells)
	if columns > numOfCells {
		row := make([]string, numOfCells)
		copy(row, cells)
		for i := 0; i < columns-numOfCells; i++ {
			row = append(row, filler)
		}
		return row
	}
	return cells
}

// calculateTableWidth calculates the width of the table.
// If the table is wider than the terminal, a table starts falling apart.
// To set a fix width could help, but in that case, even if the table is smaller, it takes the whole terminal width.
// And it doesn't look good.
func calculateTableWidth(t *ltable.Table, terminalWidth int) int {
	tableWidth := lipgloss.Width(t.String())
	if tableWidth >= terminalWidth {
		return terminalWidth
	}
	return 0
}
