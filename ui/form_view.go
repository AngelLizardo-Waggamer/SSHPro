package ui

import (
	"fmt"
	"strings"
)

func (m Model) viewForm() string {
	var b strings.Builder

	formTitle := "Nuevo host"
	if m.editIndex >= 0 {
		formTitle = "Editar host"
	}

	b.WriteString(m.styles.formTitle.Render(formTitle))
	b.WriteString("\n\n")

	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	if m.formError != "" {
		b.WriteString("\n\n")
		b.WriteString(m.styles.errorMessage.Render(fmt.Sprintf("Error: %s", m.formError)))
	}

	b.WriteString("\n\n")
	b.WriteString(m.styles.help.Render("[tab] siguiente • [shift+tab] anterior • [enter] guardar • [esc] cancelar"))

	return m.renderFrame(b.String(), m.styles.formContainer)
}
