package styles

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

const (
	// SYMBOLS
	SuccessIcon = "✔"
	ErrorIcon   = "✗"
	SelectIcon  = "➔"
	InfoIcon    = "➤"
	WarningIcon = "!"

	// ANSI color codes
	// COLORS
	prefixTextColorLight = "15"
	prefixTextColorDark  = "16"

	colorGreenLight = "28"
	colorGreenDark  = "10"

	colorYellowLight = "142"
	colorYellowDark  = "11"

	colorBlueLight = "33"
	colorBlueDark  = "45"

	colorBlack = "16"
	colorWhite = "15"

	colorRedLight = "196"
	colorRedDark  = "196"

	colorLBlueLight = "4"
	ColorLBlueDark  = "139"

	colorMagenta = "5"
	colorLBlue   = "51"
)

var defaultRender = lipgloss.NewRenderer(os.Stdout)

func DefaultStyle() lipgloss.Style {
	return defaultRender.NewStyle()
}

func DialogBox() lipgloss.Style {
	return DefaultStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(InfoStyle().GetForeground()).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)
}

func DialogButton() lipgloss.Style {
	return DefaultStyle().
		Foreground(InfoPrefix().GetForeground()).
		Background(InfoPrefix().GetBackground()).
		Padding(0, 3).
		MarginRight(2).
		MarginTop(1)
}

func ActiveDialogButton() lipgloss.Style {
	return DialogButton().
		Foreground(SelectPrefix().GetForeground()).
		Background(SelectPrefix().GetBackground())
}

func TableBorderStyle() lipgloss.Style {
	return infoStyle
}

func TableRow() lipgloss.Style {
	return infoStyle.
		PaddingLeft(1).
		PaddingRight(1)
}

func TableRowActive() lipgloss.Style {
	return selectStyle.
		Background(lipgloss.ANSIColor(240)).
		PaddingLeft(1).
		PaddingRight(1)
}

func TableRowDisabled() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.ANSIColor(245)).
		PaddingLeft(1).
		PaddingRight(1)
}

func TableRowDisabledActive() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.ANSIColor(245)).
		Background(lipgloss.ANSIColor(236)).
		PaddingLeft(1).
		PaddingRight(1)
}

var helpStyle = DefaultStyle().
	Foreground(HelpColor)

func HelpStyle() lipgloss.Style {
	return helpStyle
}
