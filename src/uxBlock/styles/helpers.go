package styles

import "github.com/charmbracelet/lipgloss"

func SuccessLine(text string) Line {
	return NewLine(SuccessText(SuccessIcon), " ", SuccessPrefix(), " ", SuccessText(text))
}

// InfoWithValue is the curried form of an info line that carries a value:
// returns a style fn suitable for uxBlock.Blocks.Print(style, value).
func InfoWithValue(label string) func(string) Line {
	return func(value string) Line {
		return NewLine(InfoText(InfoIcon), " ", InfoPrefix(), " ", InfoText(label), ": ", SelectText(value))
	}
}

func InfoLine(text string) Line {
	return NewLine(InfoText(InfoIcon), " ", InfoPrefix(), " ", InfoText(text))
}

func WarningLine(text string) Line {
	return NewLine(WarningText(WarningIcon), " ", WarningPrefix(), " ", WarningText(text))
}

func ErrorLine(text string) Line {
	return NewLine(ErrorText(ErrorIcon), " ", ErrorPrefix(), " ", ErrorText(text))
}

func SelectLine(text string) Line {
	return NewLine(SelectText(SelectIcon), " ", SelectPrefix(), " ", InfoText(text))
}

func ErrorText(text string) lipgloss.Style {
	return ErrorStyle().SetString(text)
}

func SuccessText(text string) lipgloss.Style {
	return SuccessStyle().SetString(text)
}

func WarningText(text string) lipgloss.Style {
	return WarningStyle().SetString(text)
}

func InfoText(text string) lipgloss.Style {
	return InfoStyle().SetString(text)
}

func SelectText(text string) lipgloss.Style {
	return SelectStyle().SetString(text)
}
