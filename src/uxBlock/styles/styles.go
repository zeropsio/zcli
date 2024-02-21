package styles

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

const (
	SuccessIcon = "✔"
	ErrorIcon   = "✗"
	SelectIcon  = "➔"
	InfoIcon    = "➤"
	WarningIcon = "!"
)

var defaultRender = lipgloss.NewRenderer(os.Stdout)

func defaultStyle() lipgloss.Style {
	return lipgloss.NewStyle().Renderer(defaultRender)
}

func SuccessPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "231", ANSI: "97"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
		}).
		Background(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#66bb6a", ANSI256: "114", ANSI: "32"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#66bb6a", ANSI256: "114", ANSI: "32"},
		}).
		PaddingLeft(1).PaddingRight(1).
		SetString("DONE")
}

func SuccessColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#66bb6a", ANSI256: "114", ANSI: "32"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#66bb6a", ANSI256: "114", ANSI: "32"},
		})
}

func ErrorPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "231", ANSI: "97"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
		}).
		Background(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#ff1b16", ANSI256: "196", ANSI: "91"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#ff1b16", ANSI256: "196", ANSI: "91"},
		}).
		PaddingLeft(1).PaddingRight(1).
		SetString("ERR")
}

func ErrorColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#ff1b16", ANSI256: "196", ANSI: "91"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#ff1b16", ANSI256: "196", ANSI: "91"},
		})
}

func WarningPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "231", ANSI: "97"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
		}).
		Background(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#ffa726", ANSI256: "216", ANSI: "93"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#ffa726", ANSI256: "216", ANSI: "93"},
		}).
		PaddingLeft(1).PaddingRight(1).
		SetString("WARN")
}

func WarningColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#ffa726", ANSI256: "216", ANSI: "93"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#ffa726", ANSI256: "216", ANSI: "93"},
		})
}

func InfoPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "231", ANSI: "97"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
		}).
		Background(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#e6e7ec", ANSI256: "231", ANSI: "37"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#e6e7ec", ANSI256: "231", ANSI: "37"},
		}).
		PaddingLeft(1).PaddingRight(1).
		SetString("INFO")
}

func InfoColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#e6e7ec", ANSI256: "231", ANSI: "37"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#e6e7ec", ANSI256: "231", ANSI: "37"},
		})
}

func SelectPrefix() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "231", ANSI: "97"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
		}).
		Background(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#07c", ANSI256: "27", ANSI: "30"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#07c", ANSI256: "27", ANSI: "30"},
		}).
		PaddingLeft(1).PaddingRight(1).
		SetString("SELECT")
}

func SelectColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#07c", ANSI256: "27", ANSI: "30"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#07c", ANSI256: "27", ANSI: "30"},
		})
}

func CobraSectionColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#cc0077", ANSI256: "162", ANSI: "31"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#cc0077", ANSI256: "162", ANSI: "31"},
		})
}

func CobraItemNameColor() lipgloss.Style {
	return defaultStyle().
		Foreground(lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{TrueColor: "#00ccbb", ANSI256: "44", ANSI: "36"},
			Dark:  lipgloss.CompleteColor{TrueColor: "#00ccbb", ANSI256: "44", ANSI: "36"},
		})
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
		Foreground(SelectPrefix().
			GetForeground()).
		Background(SelectPrefix().GetBackground())
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
