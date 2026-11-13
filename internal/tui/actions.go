package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/scanner"
	"space-explorer/internal/session"
)

// back returns to the disk list, keeping the finished tree for "b" to return to.
func (m *Model) back() {
	if m.mode == modeTree {
		m.mode = modeVolumes
		return
	}
	if len(m.rows) > 0 {
		m.mode = modeTree
	}
}

func (m *Model) reveal(path string) {
	if m.cfg.Reveal == nil {
		return
	}
	if err := m.cfg.Reveal(path); err != nil {
		m.status = "Cannot reveal: " + err.Error()
		return
	}
	m.status = "Opened in the file manager"
}

func (m *Model) openErrors() {
	m.errs = m.cfg.Backend.Errors()
	m.errCursor, m.errTop = 0, 0
	if m.summary.ErrorCount == 0 && len(m.errs) == 0 {
		m.status = "No unreadable paths"
		return
	}
	m.mode = modeErrors
}

// rescan re-reads the selected folder. The root is a full new scan.
func (m *Model) rescan() tea.Cmd {
	r := m.rows[m.cursor]
	if r.e.Kind != session.KindDir {
		m.status = "Select a folder to rescan"
		return nil
	}
	if m.busy {
		m.status = "A rescan is already running"
		return nil
	}
	if m.cursor == 0 {
		return m.beginScan(r.e.Path)
	}
	b := m.cfg.Backend
	m.busy, m.status, m.progress = true, "", scanner.Progress{}
	path := r.e.Path
	return func() tea.Msg {
		if err := b.Rescan(path); err != nil {
			return startErrMsg{fmt.Errorf("%w", err)}
		}
		return nil
	}
}
