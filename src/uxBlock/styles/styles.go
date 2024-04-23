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
    prefixTextColorDark = "16"

    successColorLight = "28"
    successColorDark = "10"

    warningColorLight = "142"
    warningColorDark = "11"

    selectColorLight = "33"
    selectColorDark = "45"

    infoColorLight = "16"
    infoColorDark = "15"

    errorColorLight = "196"
    errorColorDark = "196"

    cobraSectionColorLight = "4"
    cobraSectionColorDark = "139"

    cobraItemColorLight = "5"
    cobraItemColorDark = "51"
)


var defaultRender = lipgloss.NewRenderer(os.Stdout)

func defaultStyle() lipgloss.Style {
	return lipgloss.NewStyle().Renderer(defaultRender)
}

func SuccessPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: successColorLight, Dark: successColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("DONE")
}

func SuccessColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: successColorLight, Dark: successColorDark})
}

func ErrorPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: errorColorLight, Dark: errorColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("ERR")
}

func ErrorColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: errorColorLight, Dark: errorColorDark})
}

func WarningPrefix() lipgloss.Style {
	return defaultStyle().
	    Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
        Background(lipgloss.AdaptiveColor{Light: warningColorLight, Dark: warningColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("WARN")
}

func WarningColor() lipgloss.Style {
	return defaultStyle().
	    Foreground(lipgloss.AdaptiveColor{Light: warningColorLight, Dark: warningColorDark})
}

func InfoPrefix() lipgloss.Style {
	return defaultStyle().
	    Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
        Background(lipgloss.AdaptiveColor{Light: infoColorLight, Dark: infoColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("INFO")
}

func InfoColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: infoColorLight, Dark: infoColorDark})
}

func SelectPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.AdaptiveColor{Light: prefixTextColorLight, Dark: prefixTextColorDark}).
		Background(lipgloss.AdaptiveColor{Light: selectColorLight, Dark: selectColorDark}).
		PaddingLeft(1).PaddingRight(1).
		SetString("SELECT")
}

func SelectColor() lipgloss.Style {
	return defaultStyle().
	    Foreground(lipgloss.AdaptiveColor{Light: selectColorLight, Dark: selectColorDark})
}

func CobraSectionColor() lipgloss.Style {
	return defaultStyle().
        Foreground(lipgloss.AdaptiveColor{Light: cobraSectionColorLight, Dark: cobraSectionColorDark}).
        BorderBottom(true).
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.AdaptiveColor{Light: cobraSectionColorLight, Dark: cobraSectionColorDark})
}

func CobraItemNameColor() lipgloss.Style {
	return defaultStyle().
	    Foreground(lipgloss.AdaptiveColor{Light: cobraItemColorLight, Dark: cobraItemColorDark})
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
	return InfoColor()
}

func TableRow() lipgloss.Style {
	return InfoColor().
		PaddingLeft(1).
		PaddingRight(1)
}

func TableRowActive() lipgloss.Style {
	return SelectColor().
		PaddingLeft(1).
		PaddingRight(1)
}
