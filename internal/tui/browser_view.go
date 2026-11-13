package tui

import (
	"path/filepath"
	"strings"
)

func (m *Model) browseView() string {
	if len(m.br.items) == 0 {
		return dimStyle.Render("  (no sub-folders)")
	}
	end := min(len(m.br.items), m.br.top+m.bodyH())
	lines := make([]string, 0, end-m.br.top)
	for i := m.br.top; i < end; i++ {
		line := trunc(browseLabel(m.br.items[i]), m.width-2)
		if i == m.br.cursor {
			lines = append(lines, selStyle.Width(m.width).Render("> "+line))
		} else {
			lines = append(lines, "  "+styleItem(m.br.items[i], line))
		}
	}
	return strings.Join(lines, "\n")
}

func browseLabel(it item) string {
	switch {
	case it.up:
		return "../"
	case it.Link:
		return it.Name + "@"
	}
	return it.Name + "/"
}

func styleItem(it item, line string) string {
	if it.up {
		return dimStyle.Render(line)
	}
	return dirStyle.Render(line)
}

func (m *Model) browseHelp() string {
	return "Enter open  Space scan  c scan here  Bksp up  ~ home  / path  v disks  q quit"
}

func baseName(p string) string { return filepath.Base(p) }
