package ui

import (
	"fmt"
	"strings"
)

func (m Model) viewList() string {
	var b strings.Builder

	b.WriteString(m.styles.title.Render(asciiTitle))
	b.WriteString("\n")
	b.WriteString(m.styles.subtitle.Render(fmt.Sprintf("by aahl • v%s", appVersion)))
	b.WriteString("\n\n")
	b.WriteString(m.list.View())

	if m.confirmDelete {
		b.WriteString("\n")
		b.WriteString(m.styles.errorMessage.Render("¿Eliminar \"" + m.confirmName + "\"?"))
	} else if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.status.Render(m.status))
	}

	b.WriteString("\n")
	if m.confirmDelete {
		b.WriteString(m.styles.help.Render("[y] confirmar • [n] cancelar"))
	} else {
		b.WriteString(m.styles.help.Render("[enter] conectar • [/] buscar • [a] añadir • [e] editar • [d] eliminar • [t] tema • [q] salir"))
	}

	return m.renderFrame(b.String(), m.styles.container)
}

func (m Model) viewTheme() string {
	var b strings.Builder

	b.WriteString(m.styles.title.Render(asciiTitle))
	b.WriteString("\n")
	b.WriteString(m.styles.subtitle.Render(fmt.Sprintf("by aahl • v%s", appVersion)))
	b.WriteString("\n\n")
	b.WriteString(m.themeList.View())

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.status.Render(m.status))
	}

	b.WriteString("\n")
	b.WriteString(m.styles.help.Render("[enter] aplicar • [esc] volver • [q] salir"))

	return m.renderFrame(b.String(), m.styles.container)
}
