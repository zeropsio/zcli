package uxBlock

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type selectConfig struct {
	label       string
	multiSelect bool
	header      *TableRow
}

func SelectLabel(label string) SelectOption {
	return func(cfg *selectConfig) {
		cfg.label = label
	}
}

func SelectEnableMultiSelect() SelectOption {
	return func(cfg *selectConfig) {
		cfg.multiSelect = true
	}
}

func SelectTableHeader(header *TableRow) SelectOption {
	return func(cfg *selectConfig) {
		cfg.header = header
	}
}

type SelectOption = func(cfg *selectConfig)

func (b *uxBlocks) Select(ctx context.Context, tableBody *TableBody, auxOptions ...SelectOption) ([]int, error) {
	cfg := selectConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	if !b.isTerminal {
		b.PrintInfo(styles.InfoLine(cfg.label))
		return nil, errors.New(i18n.T(i18n.SelectorAllowedOnlyInTerminal))
	}

	model := &selectModel{
		cfg:       cfg,
		uxBlocks:  b,
		tableBody: tableBody,
		selected:  make(map[int]struct{}),
		width:     b.terminalWidth,
		height:    b.terminalHeight,
	}
	p := tea.NewProgram(model, tea.WithoutSignalHandler(), tea.WithContext(ctx))

	if _, err := p.Run(); err != nil {
		return nil, err
	}

	if model.canceled {
		b.ctxCancel()
		return nil, context.Canceled
	}

	sortedSelection := make([]int, 0, len(model.selected))
	for value := range model.selected {
		sortedSelection = append(sortedSelection, value)
	}
	slices.Sort(sortedSelection)

	return sortedSelection, nil
}

type selectModel struct {
	cfg       selectConfig
	uxBlocks  *uxBlocks
	tableBody *TableBody
	cursor    int
	selected  map[int]struct{}
	quiting   bool
	canceled  bool

	width  int
	height int
}

func (m *selectModel) Init() tea.Cmd {
	return nil
}

func (m *selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var keyMsg tea.KeyMsg

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// update terminal size on resize
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		// continue below
		keyMsg = msg
	default:
		return m, nil
	}

	// Make sure these keys always quiting
	//nolint:exhaustive
	switch keyMsg.Type {
	case tea.KeyCtrlC:
		m.canceled = true
		return m, tea.Quit

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}

	case tea.KeyDown:
		if m.cursor < len(m.tableBody.rows)-1 {
			m.cursor++
		}

	case tea.KeyPgUp:
		m.cursor -= 5
		if m.cursor < 0 {
			m.cursor = 0
		}

	case tea.KeyPgDown:
		m.cursor += 5
		if lastItemIndex := len(m.tableBody.rows) - 1; m.cursor > lastItemIndex {
			m.cursor = lastItemIndex
		}

	case tea.KeyEnter:
		m.quiting = true

		if !m.cfg.multiSelect {
			m.selected = make(map[int]struct{})
			m.selected[m.cursor] = struct{}{}
		}

		return m, tea.Quit
	}

	if m.cfg.multiSelect {
		if keyMsg.String() == " " {
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m *selectModel) View() string {
	if m.quiting {
		return ""
	}

	totalRows := len(m.tableBody.rows)

	// we have 1 row with "Please select xyz", so always -1, and we also need to remove 2 more rows for top and bottom border
	maxRows := m.height - 3
	// if header is present, it takes additional 2 rows (1 for text and 1 for border between header and content)
	if m.cfg.header != nil {
		maxRows -= 2
	}
	// remove 1 more row for pagination
	if totalRows > maxRows {
		maxRows -= 1
	}
	if maxRows <= 0 {
		return "Your terminal window is too small to render the table."
	}

	t := table.New().
		BorderStyle(styles.InfoColor()).
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				// first col is cursor and a pager - they can always be Active
				return styles.TableRowActive().AlignHorizontal(lipgloss.Center)
			}
			if row == m.cursor%maxRows {
				return styles.TableRowActive()
			}
			return styles.TableRow()
		})

	if m.cfg.header != nil {
		headers := make([]string, 0, len(m.cfg.header.cells)+1)
		headers = append(headers, "")
		for _, header := range m.cfg.header.cells {
			headers = append(headers, strings.ToUpper(header.Text))
		}
		t = t.Headers(headers...)
	}

	addPager := func(currentPage, maxPages int) {
		t = t.Row(fmt.Sprintf("%d/%d", currentPage, maxPages))
	}

	maxPages := int(math.Ceil(float64(totalRows) / float64(maxRows)))
	for rowIndex, row := range m.tableBody.rows {
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

		cells := make([]string, 0, len(row.cells)+1)
		icon := " "
		if rowIndex == m.cursor {
			icon = styles.SelectIcon
		}
		cells = append(cells, icon)

		for _, cell := range row.cells {
			cells = append(cells, cell.Text)
		}
		t = t.Row(cells...)

		// Add pager to the last page
		if page > 0 && rowIndex == len(m.tableBody.rows)-1 {
			addPager(page+1, maxPages)
		}
	}

	s := ""
	if m.cfg.label != "" {
		s = styles.SelectLine(m.cfg.label).String() + "\n"
	}

	t.Width(calculateTableWidth(t, m.width))

	return s + t.String()
}
