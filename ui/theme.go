package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"ssh-pro/config"
)

type themeItem struct {
	theme  config.Theme
	active bool
}

func (t themeItem) Title() string { return t.theme.Name }

func (t themeItem) Description() string {
	if t.active {
		return "Activo"
	}
	return ""
}

func (t themeItem) FilterValue() string { return t.theme.Name }

func themeItemsFromConfig(cfg config.ThemeConfig) []list.Item {
	items := make([]list.Item, 0, len(cfg.Themes))
	for _, theme := range cfg.Themes {
		items = append(items, themeItem{
			theme:  theme,
			active: theme.Name == cfg.Current,
		})
	}
	return items
}

func listDelegateFromTheme(theme config.Theme) list.DefaultDelegate {
	selectedBase := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(theme.Colors.SelectedBorder)).
		Foreground(lipgloss.Color(theme.Colors.SelectedTitle)).
		Padding(0, 0, 0, 1)
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedBase
	delegate.Styles.SelectedDesc = selectedBase.Foreground(lipgloss.Color(theme.Colors.SelectedDesc))
	return delegate
}
