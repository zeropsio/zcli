package styles

import (
	"github.com/charmbracelet/lipgloss"
)

func InfoPrefix() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: colorBlack, Dark: colorWhite}).
		PaddingLeft(1).PaddingRight(1).
		SetString("INFO")
}

func SuccessPrefix() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: colorGreenLight, Dark: colorGreenDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("DONE")
}

func SelectPrefix() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: colorBlueLight, Dark: colorBlueDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("SELECT")
}

func ErrorPrefix() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: colorRedLight, Dark: colorRedDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("ERR")
}

func WarningPrefix() lipgloss.Style {
	return DefaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: colorYellowLight, Dark: colorYellowDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("WARN")
}
