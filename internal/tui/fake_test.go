package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/session"
	"space-explorer/internal/volumes"
)

// fakeBackend serves canned listings and records calls.
type fakeBackend struct {
	lists   map[string]*session.Listing
	errs    []session.ErrorInfo
	started []string
	rescans []string
	sorts   []string
	cancels int
	failing bool
}

func (f *fakeBackend) Start(root string) error {
	if f.failing {
		return fmt.Errorf("boom")
	}
	f.started = append(f.started, root)
	return nil
}
func (f *fakeBackend) Cancel()                     { f.cancels++ }
func (f *fakeBackend) Rescan(p string) error       { f.rescans = append(f.rescans, p); return nil }
func (f *fakeBackend) Errors() []session.ErrorInfo { return f.errs }
func (f *fakeBackend) Children(p, s string) (*session.Listing, error) {
	f.sorts = append(f.sorts, s)
	if l, ok := f.lists[p]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("%s: not in scan result", p)
}

func dir(name, path string, size int64, kids bool) session.Entry {
	return session.Entry{Name: name, Path: path, Kind: session.KindDir, Size: size, Percent: float64(size), HasChildren: kids}
}

func file(name, path string, size int64) session.Entry {
	return session.Entry{Name: name, Path: path, Kind: session.KindFile, Size: size, Percent: float64(size)}
}

func newFake() *fakeBackend {
	return &fakeBackend{lists: map[string]*session.Listing{
		"/r": {Path: "/r", Size: 100, Total: 100, Entries: []session.Entry{
			dir("a", "/r/a", 60, true), dir("b", "/r/b", 30, true), file("f.bin", "/r/f.bin", 10),
			{Name: "3 other items", Path: "/r", Kind: session.KindOther, Size: 1},
		}},
		"/r/a": {Path: "/r/a", Size: 60, Total: 100, Entries: []session.Entry{
			dir("deep", "/r/a/deep", 50, true), file("x", "/r/a/x", 10),
		}},
		"/r/a/deep": {Path: "/r/a/deep", Size: 50, Total: 100, Entries: []session.Entry{file("y", "/r/a/deep/y", 50)}},
		"/r/b":      {Path: "/r/b", Size: 30, Total: 100, Entries: []session.Entry{file("z", "/r/b/z", 30)}},
	}}
}

// newTree returns a model that has finished scanning "/r".
func newTree(t *testing.T) (*Model, *fakeBackend) {
	t.Helper()
	fb := newFake()
	m := New(Config{Backend: fb, Bridge: NewBridge(), Home: "/home/u",
		Volumes: func() []volumes.Volume { return []volumes.Volume{{Name: "Disk", Path: "/", Total: 1000, Free: 400}} }})
	m.beginScan("/r")
	m.onDone(session.Summary{Root: "/r", Size: 100, Files: 5, Dirs: 3})
	return m, fb
}

func press(m *Model, keys ...string) {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "up", "down", "left", "right", "enter", "esc", "home", "end", "pgup", "pgdown":
			msg = tea.KeyMsg{Type: map[string]tea.KeyType{"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft,
				"right": tea.KeyRight, "enter": tea.KeyEnter, "esc": tea.KeyEsc, "home": tea.KeyHome,
				"end": tea.KeyEnd, "pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown}[k]}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m.Update(msg)
	}
}

func names(m *Model) string {
	var n []string
	for _, r := range m.rows {
		n = append(n, strings.Repeat(".", r.depth)+r.e.Name)
	}
	return strings.Join(n, ",")
}

func newTreeVolumes() []volumes.Volume {
	return []volumes.Volume{{Name: "Disk", Path: "/", Total: 1000, Free: 400}}
}
func keyEnter() tea.KeyMsg      { return tea.KeyMsg{Type: tea.KeyEnter} }
func keyEsc() tea.KeyMsg        { return tea.KeyMsg{Type: tea.KeyEsc} }
func keyRune(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }
