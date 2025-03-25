package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var infoColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: infoColorLight, Dark: infoColorDark})

func InfoColor() lipgloss.Style {
	return infoColor
}

var successColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: successColorLight, Dark: successColorDark})

func SuccessColor() lipgloss.Style {
	return successColor
}

var selectColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: selectColorLight, Dark: selectColorDark})

func SelectColor() lipgloss.Style {
	return selectColor
}

var warningColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: warningColorLight, Dark: warningColorDark})

func WarningColor() lipgloss.Style {
	return warningColor
}

var errorColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: errorColorLight, Dark: errorColorDark})

func ErrorColor() lipgloss.Style {
	return errorColor
}

var cobraItemNameColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: cobraItemColorLight, Dark: cobraItemColorDark})

func CobraItemNameColor() lipgloss.Style {
	return cobraItemNameColor
}

var cobraSelectionColor = defaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: cobraSectionColorLight, Dark: cobraSectionColorDark}).
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.AdaptiveColor{Light: cobraSectionColorLight, Dark: cobraSectionColorDark})

func CobraSectionColor() lipgloss.Style {
	return cobraSelectionColor
}
