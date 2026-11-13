package tui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) volumeKey(k tea.KeyMsg) tea.Cmd {
	m.status = ""
	switch k.String() {
	case "q", "esc":
		return m.quitCmd()
	case "up", "k":
		m.volCursor = max(0, m.volCursor-1)
	case "down", "j":
		m.volCursor = min(len(m.vols)-1, m.volCursor+1)
	case "enter", " ":
		if len(m.vols) > 0 {
			return m.beginScan(m.vols[m.volCursor].Path)
		}
	case "h":
		if m.cfg.Home != "" {
			return m.beginScan(m.cfg.Home)
		}
	case "/", "p":
		m.openPrompt("")
	case "f", "tab", "right":
		if len(m.vols) > 0 {
			m.openBrowser(m.vols[m.volCursor].Path, "")
		}
	case "b":
		m.back()
	}
	return nil
}

func (m *Model) promptKey(k tea.KeyMsg) tea.Cmd {
	switch k.Type {
	case tea.KeyEsc:
		m.mode, m.status = m.promptFrom, ""
	case tea.KeyEnter:
		path := m.expandHome(strings.TrimSpace(m.input))
		if path == "" {
			return nil
		}
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			m.status = path + ": not a folder"
			return nil
		}
		return m.beginScan(path)
	case tea.KeyBackspace:
		if r := []rune(m.input); len(r) > 0 {
			m.input = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		m.input += " "
	case tea.KeyRunes:
		m.input += string(k.Runes)
	}
	return nil
}

// openPrompt asks for a path; start prefills it (the folder being browsed).
func (m *Model) openPrompt(start string) {
	m.promptFrom, m.mode, m.input = m.mode, modePrompt, ""
	if start != "" {
		m.input = strings.TrimSuffix(start, string(os.PathSeparator)) + string(os.PathSeparator)
	}
}

func (m *Model) expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		return m.cfg.Home + strings.TrimPrefix(p, "~")
	}
	return p
}

func (m *Model) scanningKey(k tea.KeyMsg) tea.Cmd {
	if k.String() == "esc" || k.String() == "q" {
		return m.cancelCmd()
	}
	return nil
}

func (m *Model) errorsKey(k tea.KeyMsg) tea.Cmd {
	switch k.String() {
	case "esc", "q", "e":
		m.mode = modeTree
	case "up", "k":
		m.moveErr(m.errCursor - 1)
	case "down", "j":
		m.moveErr(m.errCursor + 1)
	case "pgup":
		m.moveErr(m.errCursor - m.bodyH())
	case "pgdown":
		m.moveErr(m.errCursor + m.bodyH())
	case "enter", "o":
		if len(m.errs) > 0 {
			m.reveal(m.errs[m.errCursor].Path)
		}
	}
	return nil
}

func (m *Model) moveErr(i int) {
	m.errCursor = max(0, min(i, len(m.errs)-1))
	h := m.bodyH()
	m.errTop = max(0, min(m.errTop, m.errCursor))
	if m.errCursor >= m.errTop+h {
		m.errTop = m.errCursor - h + 1
	}
}
