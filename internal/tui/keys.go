package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/session"
)

func (m *Model) onKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if k.String() == "ctrl+c" {
		return m, m.quitCmd()
	}
	switch m.mode {
	case modeVolumes:
		return m, m.volumeKey(k)
	case modePrompt:
		return m, m.promptKey(k)
	case modeScanning:
		return m, m.scanningKey(k)
	case modeErrors:
		return m, m.errorsKey(k)
	case modeBrowse:
		return m, m.browseKey(k)
	}
	return m, m.treeKey(k)
}

func (m *Model) treeKey(k tea.KeyMsg) tea.Cmd {
	if len(m.rows) == 0 {
		return nil
	}
	m.status = ""
	switch k.String() {
	case "q":
		return m.quitCmd()
	case "esc":
		if m.busy {
			return m.cancelCmd()
		}
	case "up", "k":
		m.moveTo(m.cursor - 1)
	case "down", "j":
		m.moveTo(m.cursor + 1)
	case "pgup":
		m.moveTo(m.cursor - m.bodyH())
	case "pgdown":
		m.moveTo(m.cursor + m.bodyH())
	case "home", "g":
		m.moveTo(0)
	case "end", "G":
		m.moveTo(len(m.rows) - 1)
	case "enter", " ":
		m.toggle()
	case "right", "l":
		m.expandOrDescend()
	case "left", "h":
		m.collapseOrAscend()
	case "s":
		m.toggleSort()
	case "r":
		return m.rescan()
	case "o":
		m.reveal(m.rows[m.cursor].e.Path)
	case "e":
		m.openErrors()
	case "b":
		m.back()
	}
	return nil
}

func (m *Model) moveTo(i int) {
	m.cursor = max(0, min(i, len(m.rows)-1))
	m.fixScroll()
}

func (m *Model) toggle() {
	r := m.rows[m.cursor]
	switch {
	case !r.expandable():
	case r.expanded:
		m.rows = collapseAt(m.rows, m.cursor)
	default:
		m.expand()
	}
}

func (m *Model) expand() {
	r := m.rows[m.cursor]
	l, err := m.fetch(r.e.Path)
	if err != nil {
		m.status = err.Error()
		return
	}
	m.rows[m.cursor].expanded = true
	m.rows = insertAfter(m.rows, m.cursor, rowsOf(l, r.depth+1))
}

func (m *Model) expandOrDescend() {
	r := m.rows[m.cursor]
	switch {
	case r.expandable() && !r.expanded:
		m.expand()
	case r.expanded:
		m.moveTo(m.cursor + 1)
	}
}

func (m *Model) collapseOrAscend() {
	if m.rows[m.cursor].expanded {
		m.rows = collapseAt(m.rows, m.cursor)
	} else if p := parentIndex(m.rows, m.cursor); p >= 0 {
		m.moveTo(p)
	}
	m.fixScroll()
}

func (m *Model) toggleSort() {
	if m.sortBy == session.SortName {
		m.sortBy, m.status = session.SortSize, "Sorted by size"
	} else {
		m.sortBy, m.status = session.SortName, "Sorted by name"
	}
	m.refresh()
}
