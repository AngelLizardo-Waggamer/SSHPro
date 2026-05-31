package ui

import "strings"

func (m Model) viewList() string {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("Hosts SSH"))
	b.WriteString("\n\n")
	b.WriteString(m.list.View())

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.status.Render(m.status))
	}

	b.WriteString("\n")
	b.WriteString(m.styles.help.Render("[enter] conectar • [a] añadir • [e] editar • [d] eliminar • [q] salir"))

	return m.styles.container.Render(b.String())
}
