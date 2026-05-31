package ui

import "strings"

func (m Model) viewList() string {
	var b strings.Builder

	b.WriteString(m.styles.title.Render(asciiTitle))
	b.WriteString("\n")
	b.WriteString(m.styles.subtitle.Render("by aahl"))
	b.WriteString("\n\n")
	b.WriteString(m.list.View())

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.status.Render(m.status))
	}

	b.WriteString("\n")
	b.WriteString(m.styles.help.Render("[enter] conectar • [/] buscar • [a] añadir • [e] editar • [d] eliminar • [q] salir"))

	return m.renderFrame(b.String(), m.styles.container)
}
