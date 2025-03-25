package styles

import (
	"github.com/charmbracelet/lipgloss"
)

func InfoPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: infoColorLight, Dark: infoColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("INFO")
}

func SuccessPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: successColorLight, Dark: successColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("DONE")
}

func SelectPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: selectColorLight, Dark: selectColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("SELECT")
}

func ErrorPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: errorColorLight, Dark: errorColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("ERR")
}

func WarningPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: warningColorLight, Dark: warningColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("WARN")
}
