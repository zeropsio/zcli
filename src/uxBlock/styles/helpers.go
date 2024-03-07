package styles

import "github.com/charmbracelet/lipgloss"

func SuccessLine(text string) Line {
	return NewLine(SuccessText(SuccessIcon), " ", SuccessPrefix(), " ", SuccessText(text))
}

func InfoWithValueLine(text string, value string) Line {
	return NewLine(InfoText(InfoIcon), " ", InfoPrefix(), " ", InfoText(text), ": ", SelectText(value))
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
	return ErrorColor().SetString(text)
}

func SuccessText(text string) lipgloss.Style {
	return SuccessColor().SetString(text)
}

func WarningText(text string) lipgloss.Style {
	return WarningColor().SetString(text)
}

func InfoText(text string) lipgloss.Style {
	return InfoColor().SetString(text)
}

func SelectText(text string) lipgloss.Style {
	return SelectColor().SetString(text)
}
