package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ssh-pro/config"
	"ssh-pro/ssh"
	"ssh-pro/storage"
	"ssh-pro/ui"
)

func main() {
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

	if len(os.Args) > 1 {
		query := strings.Join(os.Args[1:], " ")
		matches := storage.FindHostsByName(hosts, query)
		if len(matches) == 0 {
			fmt.Println("No se encontró ningún servidor que coincida con el nombre especificado")
			return
		}
		if len(matches) > 1 {
			fmt.Println("Se encontraron 2 o más coincidencias de nombre:")
			for _, host := range matches {
				fmt.Printf("* %s\n", host.Name)
			}
			fmt.Println("Intenta con un nombre más específico")
			return
		}

		cmd, err := ssh.BuildCommand(matches[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error al preparar SSH: %v\n", err)
			os.Exit(1)
		}
		if err := cmd.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error al ejecutar SSH: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if _, err := config.EnsureThemeFile(); err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo preparar el tema: %v\n", err)
		os.Exit(1)
	}

	themeConfig, err := config.LoadThemeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo cargar el tema: %v\n", err)
		os.Exit(1)
	}

	model := ui.NewModel(store, hosts, themeConfig)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error al ejecutar la UI: %v\n", err)
		os.Exit(1)
	}
}
