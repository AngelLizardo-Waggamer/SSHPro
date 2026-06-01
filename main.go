package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"ssh-pro/config"
	"ssh-pro/storage"
	"ssh-pro/ui"
)

func main() {
	if _, err := config.EnsureThemeFile(); err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo preparar el tema: %v\n", err)
		os.Exit(1)
	}

	themeConfig, err := config.LoadThemeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo cargar el tema: %v\n", err)
		os.Exit(1)
	}

	path, err := storage.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo resolver la ruta de configuración: %v\n", err)
		os.Exit(1)
	}

	store := storage.NewStore(path)
	hosts, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo cargar la configuración: %v\n", err)
		os.Exit(1)
	}

	model := ui.NewModel(store, hosts, themeConfig)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error al ejecutar la UI: %v\n", err)
		os.Exit(1)
	}
}
