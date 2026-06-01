package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Theme defines a color scheme for the UI.
type Theme struct {
	Name   string      `json:"name"`
	Colors ThemeColors `json:"colors"`
}

// ThemeConfig stores the active theme and available themes.
type ThemeConfig struct {
	Current string  `json:"current"`
	Themes  []Theme `json:"themes"`
}

// ThemeColors contains the color tokens used by the UI.
type ThemeColors struct {
	ContainerBorder  string `json:"container_border"`
	Title            string `json:"title"`
	Help             string `json:"help"`
	Status           string `json:"status"`
	Subtitle         string `json:"subtitle"`
	FocusedInput     string `json:"focused_input"`
	BlurredInput     string `json:"blurred_input"`
	ErrorMessage     string `json:"error_message"`
	FormTitle        string `json:"form_title"`
	FormBorder       string `json:"form_border"`
	SelectedBorder   string `json:"selected_border"`
	SelectedTitle    string `json:"selected_title"`
	SelectedDesc     string `json:"selected_desc"`
	FilterPrompt     string `json:"filter_prompt"`
	FilterText       string `json:"filter_text"`
}

// DefaultTheme returns the current blue theme.
func DefaultTheme() Theme {
	return Theme{
		Name: "blue",
		Colors: ThemeColors{
			ContainerBorder: "33",
			Title:           "39",
			Help:            "110",
			Status:          "81",
			Subtitle:        "244",
			FocusedInput:    "81",
			BlurredInput:    "250",
			ErrorMessage:    "203",
			FormTitle:       "39",
			FormBorder:      "33",
			SelectedBorder:  "33",
			SelectedTitle:   "81",
			SelectedDesc:    "75",
			FilterPrompt:    "75",
			FilterText:      "81",
		},
	}
}

// DefaultThemeConfig returns the default config with the blue theme.
func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Current: "blue",
		Themes:  []Theme{DefaultTheme()},
	}
}

// DefaultThemePath returns the default theme JSON path under ~/.sshpro.
func DefaultThemePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".sshpro", "themes.json"), nil
}

// EnsureThemeFile creates the theme file with the default theme if it doesn't exist.
func EnsureThemeFile() (string, error) {
	path, err := DefaultThemePath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat theme file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create theme dir: %w", err)
	}

	data, err := json.MarshalIndent(DefaultThemeConfig(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode theme JSON: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write theme JSON: %w", err)
	}

	return path, nil
}

// LoadThemeConfig reads the theme configuration from disk.
func LoadThemeConfig() (ThemeConfig, error) {
	path, err := DefaultThemePath()
	if err != nil {
		return ThemeConfig{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ThemeConfig{}, fmt.Errorf("read theme JSON: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return DefaultThemeConfig(), nil
	}

	var cfg ThemeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ThemeConfig{}, fmt.Errorf("parse theme JSON: %w", err)
	}

	if len(cfg.Themes) == 0 {
		return DefaultThemeConfig(), nil
	}

	return cfg, nil
}

// SaveThemeConfig writes the theme configuration to disk.
func SaveThemeConfig(cfg ThemeConfig) error {
	path, err := DefaultThemePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create theme dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode theme JSON: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write theme JSON: %w", err)
	}
	return nil
}

// FindTheme finds a theme by name.
func FindTheme(cfg ThemeConfig, name string) (Theme, bool) {
	for _, theme := range cfg.Themes {
		if theme.Name == name {
			return theme, true
		}
	}
	return Theme{}, false
}
