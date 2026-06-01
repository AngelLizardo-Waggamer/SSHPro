package ui

import (
	"github.com/charmbracelet/lipgloss"

	"ssh-pro/config"
)

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

func stylesFromTheme(theme config.Theme) styles {
	base := lipgloss.NewStyle().Padding(1, 2)
	colors := theme.Colors
	return styles{
		container:     base.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colors.ContainerBorder)),
		title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Title)),
		help:          lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Help)),
		status:        lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Status)),
		subtitle:      lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Subtitle)),
		focusedInput:  lipgloss.NewStyle().Foreground(lipgloss.Color(colors.FocusedInput)),
		blurredInput:  lipgloss.NewStyle().Foreground(lipgloss.Color(colors.BlurredInput)),
		errorMessage:  lipgloss.NewStyle().Foreground(lipgloss.Color(colors.ErrorMessage)),
		formTitle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.FormTitle)),
		formContainer: base.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colors.FormBorder)),
	}
}
