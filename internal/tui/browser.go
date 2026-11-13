package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/browse"
)

// item is one line of the folder browser; up is the ".." line.
type item struct {
	browse.Entry
	up bool
}

// browser is the state of the folder picker (a file-manager style list).
type browser struct {
	dir    string
	items  []item
	cursor int
	top    int
}

// openBrowser shows dir. If it cannot be read the current screen stays and
// the reason goes to the status line. keep names the entry to highlight
// (the folder we just came out of).
func (m *Model) openBrowser(dir, keep string) {
	entries, err := browse.List(dir)
	if err != nil {
		m.status = "Cannot open " + dir + ": " + reason(err)
		return
	}
	items := make([]item, 0, len(entries)+1)
	if _, ok := browse.Parent(dir); ok {
		items = append(items, item{up: true, Entry: browse.Entry{Name: "..", Path: dir}})
	}
	for _, e := range entries {
		items = append(items, item{Entry: e})
	}
	m.br = browser{dir: dir, items: items}
	for i, it := range items {
		if !it.up && it.Name == keep {
			m.br.cursor = i
		}
	}
	m.mode, m.status = modeBrowse, ""
	m.fixBrowseScroll()
}

func reason(err error) string {
	if pe, ok := err.(*os.PathError); ok {
		return pe.Err.Error()
	}
	return err.Error()
}

func (m *Model) browseKey(k tea.KeyMsg) tea.Cmd {
	m.status = ""
	switch k.String() {
	case "q":
		return m.quitCmd()
	case "esc", "v", "tab":
		m.mode = modeVolumes
	case "up", "k":
		m.moveBrowse(m.br.cursor - 1)
	case "down", "j":
		m.moveBrowse(m.br.cursor + 1)
	case "pgup":
		m.moveBrowse(m.br.cursor - m.bodyH())
	case "pgdown":
		m.moveBrowse(m.br.cursor + m.bodyH())
	case "home", "g":
		m.moveBrowse(0)
	case "end", "G":
		m.moveBrowse(len(m.br.items) - 1)
	case "enter", "right", "l":
		m.browseOpen()
	case "left", "backspace", "h":
		m.browseUp()
	case "~":
		if m.cfg.Home != "" {
			m.openBrowser(m.cfg.Home, "")
		}
	case " ", "s":
		return m.browseScanSelected()
	case "c":
		return m.browseScan(m.br.dir)
	case "/":
		m.openPrompt(m.br.dir)
	}
	return nil
}

func (m *Model) selected() (item, bool) {
	if m.br.cursor < 0 || m.br.cursor >= len(m.br.items) {
		return item{}, false
	}
	return m.br.items[m.br.cursor], true
}

func (m *Model) browseOpen() {
	it, ok := m.selected()
	switch {
	case !ok:
	case it.up:
		m.browseUp()
	default:
		m.openBrowser(it.Path, "")
	}
}

func (m *Model) browseUp() {
	if p, ok := browse.Parent(m.br.dir); ok {
		m.openBrowser(p, baseName(m.br.dir))
	}
}

func (m *Model) browseScanSelected() tea.Cmd {
	it, ok := m.selected()
	switch {
	case !ok:
		return nil
	case it.up:
		m.status = "Pick a folder, or press c to scan " + m.br.dir
		return nil
	}
	return m.browseScan(it.Path)
}

// browseScan scans path; cancelling or failing returns to the browser.
func (m *Model) browseScan(path string) tea.Cmd {
	cmd := m.beginScan(path)
	m.cancelTo = modeBrowse
	return cmd
}

func (m *Model) moveBrowse(i int) {
	m.br.cursor = max(0, min(i, len(m.br.items)-1))
	m.fixBrowseScroll()
}

func (m *Model) fixBrowseScroll() {
	h := m.bodyH()
	if m.br.cursor < m.br.top {
		m.br.top = m.br.cursor
	}
	if m.br.cursor >= m.br.top+h {
		m.br.top = m.br.cursor - h + 1
	}
	m.br.top = max(0, min(m.br.top, len(m.br.items)-h))
}

func (m *Model) browseStatus() string {
	n := len(m.br.items)
	if n > 0 && m.br.items[0].up {
		n--
	}
	return fmt.Sprintf("%s folders", formatCount(int64(n)))
}
