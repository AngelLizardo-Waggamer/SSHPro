package ui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	container     lipgloss.Style
	title         lipgloss.Style
	help          lipgloss.Style
	status        lipgloss.Style
	focusedInput  lipgloss.Style
	blurredInput  lipgloss.Style
	errorMessage  lipgloss.Style
	formTitle     lipgloss.Style
	formContainer lipgloss.Style
}

func defaultStyles() styles {
	base := lipgloss.NewStyle().Padding(1, 2)
	return styles{
		container:     base.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")),
		title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		help:          lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		status:        lipgloss.NewStyle().Foreground(lipgloss.Color("84")),
		focusedInput:  lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		blurredInput:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		errorMessage:  lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		formTitle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141")),
		formContainer: base.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("99")),
	}
}
