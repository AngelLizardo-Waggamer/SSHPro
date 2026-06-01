package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	frameHorizontal   = 6 // border (2) + padding (4)
	frameVertical     = 4 // border (2) + padding (2)
	listFooterLines   = 3 // blank + status + help
	listHeaderPadding = 2 // subtitle + blank line
)

const appVersion = "1.1"

const asciiTitle = `::::::::   ::::::::  :::    ::: :::::::::  :::::::::   ::::::::
:+:    :+: :+:    :+: :+:    :+: :+:    :+: :+:    :+: :+:    :+:
+:+        +:+        +:+    +:+ +:+    +:+ +:+    +:+ +:+    +:+
+#++:++#++ +#++:++#++ +#++:++#++ +#++:++#+  +#++:++#:  +#+    +:+
      +#+        +#+ +#+    +#+ +#+        +#+    +#+ +#+    +#+
#+#    #+# #+#    #+# #+#    #+# #+#        #+#    #+# #+#    #+#
########   ########  ###    ### ###        ###    ###  ########`

func listReservedLines() int {
	titleLines := strings.Count(asciiTitle, "\n") + 1
	return titleLines + listHeaderPadding + listFooterLines
}

func (m Model) contentSize() (int, int) {
	width := m.width - frameHorizontal
	height := m.height - frameVertical
	return max(0, width), max(0, height)
}

func (m Model) renderFrame(content string, style lipgloss.Style) string {
	width, height := m.contentSize()
	padded := padContent(content, width, height)
	return style.Render(padded)
}

func padContent(content string, width, height int) string {
	if width <= 0 && height <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	if width > 0 {
		for i, line := range lines {
			lineWidth := lipgloss.Width(line)
			if lineWidth < width {
				lines[i] = line + strings.Repeat(" ", width-lineWidth)
			}
		}
	}

	if height > 0 && len(lines) < height {
		padLine := ""
		if width > 0 {
			padLine = strings.Repeat(" ", width)
		}
		for len(lines) < height {
			lines = append(lines, padLine)
		}
	}

	return strings.Join(lines, "\n")
}
