package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var ContrastColor = lipgloss.AdaptiveColor{Light: colorBlack, Dark: colorWhite}
var InfoColor = lipgloss.AdaptiveColor{Light: colorBlack, Dark: colorWhite}
var infoStyle = DefaultStyle().
	Foreground(InfoColor)

func InfoStyle() lipgloss.Style { return infoStyle }

var GreenColor = lipgloss.AdaptiveColor{Light: colorGreenLight, Dark: colorGreenDark}
var SuccessColor = lipgloss.AdaptiveColor{Light: colorGreenLight, Dark: colorGreenDark}
var successStyle = DefaultStyle().
	Foreground(SuccessColor)

func SuccessStyle() lipgloss.Style {
	return successStyle
}

var BlueColor = lipgloss.AdaptiveColor{Light: colorBlueLight, Dark: colorBlueDark}
var SelectColor = lipgloss.AdaptiveColor{Light: colorBlueLight, Dark: colorBlueDark}
var selectStyle = DefaultStyle().
	Foreground(SelectColor)

func SelectStyle() lipgloss.Style {
	return selectStyle
}

var YellowColor = lipgloss.AdaptiveColor{Light: colorYellowLight, Dark: colorYellowDark}
var WarningColor = lipgloss.AdaptiveColor{Light: colorYellowLight, Dark: colorYellowDark}
var warningStyle = DefaultStyle().
	Foreground(WarningColor)

func WarningStyle() lipgloss.Style {
	return warningStyle
}

var RedColor = lipgloss.AdaptiveColor{Light: colorRedLight, Dark: colorRedDark}
var ErrorColor = lipgloss.AdaptiveColor{Light: colorRedLight, Dark: colorRedDark}
var errorStyle = DefaultStyle().
	Foreground(ErrorColor)

func ErrorStyle() lipgloss.Style {
	return errorStyle
}

var cobraItemNameStyle = DefaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: colorMagenta, Dark: colorLBlue})

func CobraItemNameStyle() lipgloss.Style {
	return cobraItemNameStyle
}

var cobraSelectionStyle = DefaultStyle().
	Foreground(lipgloss.AdaptiveColor{Light: colorLBlueLight, Dark: ColorLBlueDark}).
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.AdaptiveColor{Light: colorLBlueLight, Dark: ColorLBlueDark})

func CobraSectionStyle() lipgloss.Style {
	return cobraSelectionStyle
}

var HelpColor = lipgloss.AdaptiveColor{
	Light: "242",
	Dark:  "246",
}
