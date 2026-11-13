package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/scanner"
	"space-explorer/internal/session"
	"space-explorer/internal/volumes"
)

type mode int

const (
	modeVolumes mode = iota
	modePrompt
	modeScanning
	modeTree
	modeErrors
	modeBrowse
)

// Model is the whole UI state. All methods use a pointer receiver.
type Model struct {
	cfg           Config
	mode          mode
	width, height int

	vols       []volumes.Volume
	volCursor  int
	input      string // path prompt text
	promptFrom mode   // screen to return to when the prompt is cancelled
	cancelTo   mode   // screen shown if a scan is cancelled or fails
	br         browser

	root     string
	progress scanner.Progress
	frame    int  // spinner frame, advances with progress events
	busy     bool // a sub-folder rescan runs while the tree stays browsable

	rows    []row
	cursor  int
	top     int // first visible row
	sortBy  string
	summary session.Summary
	status  string

	errs      []session.ErrorInfo
	errCursor int
	errTop    int
}

// startErrMsg reports that Backend.Start refused to begin.
type startErrMsg struct{ err error }

// New builds the model. If cfg.Start is set the scan begins immediately.
func New(cfg Config) *Model {
	m := &Model{cfg: cfg, width: 80, height: 24, sortBy: session.SortSize}
	if cfg.Volumes != nil {
		m.vols = cfg.Volumes()
	}
	if cfg.Start != "" {
		m.mode, m.root = modeScanning, cfg.Start
	}
	return m
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.cfg.Bridge.wait()}
	if m.mode == modeScanning {
		cmds = append(cmds, m.startCmd(m.root))
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.fixScroll()
	case eventMsg:
		return m, m.onEvent(msg)
	case startErrMsg:
		m.onStopped("Cannot scan: " + msg.err.Error())
	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return m, nil
}

func (m *Model) startCmd(root string) tea.Cmd {
	b := m.cfg.Backend
	return func() tea.Msg {
		if err := b.Start(root); err != nil {
			return startErrMsg{err}
		}
		return nil
	}
}

func (m *Model) cancelCmd() tea.Cmd {
	b := m.cfg.Backend
	m.status = "Cancelling..."
	return func() tea.Msg { b.Cancel(); return nil }
}

func (m *Model) quitCmd() tea.Cmd {
	b := m.cfg.Backend
	return func() tea.Msg { b.Cancel(); return tea.Quit() }
}

// beginScan switches to the progress screen; Backend.Start drops the old
// result, so the rows of the previous tree are dropped too.
func (m *Model) beginScan(root string) tea.Cmd {
	m.root, m.rows, m.cursor, m.top = root, nil, 0, 0
	m.progress, m.status, m.mode, m.cancelTo = scanner.Progress{}, "", modeScanning, modeVolumes
	return m.startCmd(root)
}

func (m *Model) fetch(path string) (*session.Listing, error) {
	return m.cfg.Backend.Children(path, m.sortBy)
}

// load shows a freshly finished scan.
func (m *Model) load() {
	rows, err := loadTree(m.fetch, m.root, nil)
	if err != nil {
		m.status, m.mode = err.Error(), modeVolumes
		return
	}
	m.rows, m.cursor, m.top, m.mode = rows, 0, 0, modeTree
}

// refresh reloads the tree keeping expanded folders and the selection.
func (m *Model) refresh() {
	if len(m.rows) == 0 {
		return
	}
	key := m.rows[m.cursor].key()
	rows, err := loadTree(m.fetch, m.rows[0].e.Path, openSet(m.rows))
	if err != nil {
		m.status = err.Error()
		return
	}
	m.rows = rows
	if i := indexOfKey(rows, key); i >= 0 {
		m.cursor = i
	} else {
		m.cursor = min(m.cursor, len(rows)-1)
	}
	m.fixScroll()
}

func (m *Model) bodyH() int { return max(1, m.height-4) }

func (m *Model) fixScroll() {
	h := m.bodyH()
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+h {
		m.top = m.cursor - h + 1
	}
	m.top = max(0, min(m.top, len(m.rows)-h))
}
