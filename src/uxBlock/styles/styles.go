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

	// COLORS
	prefixTextColorLight = "15"
	prefixTextColorDark  = "16"

	successColorLight = "28"
	successColorDark  = "10"

	warningColorLight = "142"
	warningColorDark  = "11"

	selectColorLight = "33"
	selectColorDark  = "45"

	infoColorLight = "16"
	infoColorDark  = "15"

	errorColorLight = "196"
	errorColorDark  = "196"

	cobraSectionColorLight = "4"
	cobraSectionColorDark  = "139"

	cobraItemColorLight = "5"
	cobraItemColorDark  = "51"
)

var defaultRender = lipgloss.NewRenderer(os.Stdout)

func defaultStyle() lipgloss.Style {
	return lipgloss.NewStyle().Renderer(defaultRender)
}

func DialogBox() lipgloss.Style {
	return defaultStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(InfoColor().GetForeground()).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)
}

func DialogButton() lipgloss.Style {
	return defaultStyle().
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
	return infoColor
}

func TableRow() lipgloss.Style {
	return infoColor.
		PaddingLeft(1).
		PaddingRight(1)
}

func TableRowActive() lipgloss.Style {
	return selectColor.
		PaddingLeft(1).
		PaddingRight(1)
}
