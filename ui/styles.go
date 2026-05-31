package ui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	container     lipgloss.Style
	title         lipgloss.Style
	help          lipgloss.Style
	status        lipgloss.Style
	subtitle      lipgloss.Style
	focusedInput  lipgloss.Style
	blurredInput  lipgloss.Style
	errorMessage  lipgloss.Style
	formTitle     lipgloss.Style
	formContainer lipgloss.Style
}

func defaultStyles() styles {
	base := lipgloss.NewStyle().Padding(1, 2)
	return styles{
		container:     base.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("33")),
		title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
		help:          lipgloss.NewStyle().Foreground(lipgloss.Color("110")),
		status:        lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		subtitle:      lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		focusedInput:  lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		blurredInput:  lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		errorMessage:  lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		formTitle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
		formContainer: base.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("33")),
	}
}
