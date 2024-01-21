package uxBlock

import (
	"context"
	"errors"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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

func (b *UxBlocks) Select(ctx context.Context, tableBody *TableBody, auxOptions ...SelectOption) ([]int, error) {
	cfg := selectConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	// FIXME - janhajek fix message
	if !b.isTerminal {
		return nil, errors.New(cfg.label + ", you can choose only in terminal")
	}

	model := &selectModel{
		cfg:       cfg,
		uxBlocks:  b,
		tableBody: tableBody,
		selected:  make(map[int]struct{}),
	}
	p := tea.NewProgram(model, tea.WithoutSignalHandler(), tea.WithContext(ctx))
	_, err := p.Run()
	if err != nil {
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
	uxBlocks  *UxBlocks
	tableBody *TableBody
	cursor    int
	selected  map[int]struct{}
	quiting   bool
	canceled  bool
}

func (m *selectModel) Init() tea.Cmd {
	return nil
}

func (m *selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quiting
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {

		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			if m.cursor < len(m.tableBody.rows)-1 {
				m.cursor++
			}
		case "enter":
			m.quiting = true

			if !m.cfg.multiSelect {
				m.selected = make(map[int]struct{})
				m.selected[m.cursor] = struct{}{}
			}

			return m, tea.Quit
		}

		if m.cfg.multiSelect {
			switch msg.String() {

			case " ":
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}
		}
	}

	return m, nil
}

func (m *selectModel) View() string {
	if m.quiting {
		return ""
	}

	baseStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			even := row%2 == 0

			if even {
				return baseStyle.Copy().Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Copy().Foreground(lipgloss.Color("252"))
		})

	if m.cfg.header != nil {
		capitalizeHeaders := func(data []string) []string {
			for i := range data {
				data[i] = strings.ToUpper(data[i])
			}
			return data
		}

		headers := make([]string, len(m.cfg.header.cells)+1)
		headers[0] = ""
		for i, header := range m.cfg.header.cells {
			headers[i+1] = header.Text
		}
		t = t.Headers(capitalizeHeaders(headers)...)
	}

	rows := make([][]string, len(m.tableBody.rows))
	for rowIndex, row := range m.tableBody.rows {
		cells := make([]string, len(row.cells)+1)
		cells[0] = " "
		if rowIndex == m.cursor {
			cells[0] = "✓"
		}
		for i, cell := range row.cells {
			cells[i+1] = cell.Text
		}
		rows[rowIndex] = cells
	}
	t = t.Rows(rows...)

	s := ""
	if m.cfg.label != "" {
		s = SelectionIcon + m.cfg.label + "\n"
	}

	return s + t.String()
}
