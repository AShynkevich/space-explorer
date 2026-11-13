package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBrowseUnreadableFolderKeepsPlaceAndExplains(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read everything")
	}
	m, _, root := browseModel(t)
	locked := filepath.Join(root, "beta")
	_ = os.Chmod(locked, 0)
	defer os.Chmod(locked, 0o755)
	press(m, "f", "down", "down", "enter")
	if m.mode != modeBrowse || m.br.dir != root || !strings.Contains(m.status, "Cannot open") {
		t.Fatalf("dir=%q status=%q", m.br.dir, m.status)
	}
}

func TestBrowseDeletedFolderIsReported(t *testing.T) {
	m, _, root := browseModel(t)
	press(m, "f", "down", "down")
	_ = os.RemoveAll(filepath.Join(root, "beta"))
	press(m, "enter")
	if m.br.dir != root || !strings.Contains(m.status, "Cannot open") {
		t.Fatalf("dir=%q status=%q", m.br.dir, m.status)
	}
}

func TestBrowseHomeAndEscAndPromptRoundTrip(t *testing.T) {
	m, _, root := browseModel(t)
	press(m, "f", "down", "enter", "~")
	if m.br.dir != root {
		t.Fatalf("~ went to %q", m.br.dir)
	}
	press(m, "/")
	if m.mode != modePrompt || m.input != root+string(os.PathSeparator) {
		t.Fatalf("mode=%v input=%q", m.mode, m.input)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeBrowse {
		t.Fatalf("esc from prompt -> %v, want browser", m.mode)
	}
	press(m, "esc")
	if m.mode != modeVolumes {
		t.Fatalf("esc from browser -> %v", m.mode)
	}
}

func TestBrowseScrollsAndFitsTheScreen(t *testing.T) {
	m, _, root := browseModel(t)
	for i := 0; i < 40; i++ {
		_ = os.Mkdir(filepath.Join(root, "zdir"+string(rune('a'+i%26))+string(rune('a'+i/26))), 0o755)
	}
	press(m, "f")
	m.width, m.height = 50, 10
	m.fixBrowseScroll()
	press(m, "end")
	if m.br.cursor != len(m.br.items)-1 || m.br.top != len(m.br.items)-m.bodyH() {
		t.Fatalf("cursor=%d top=%d of %d", m.br.cursor, m.br.top, len(m.br.items))
	}
	lines := strings.Split(m.View(), "\n")
	if len(lines) > 10 || !strings.Contains(m.View(), m.br.items[m.br.cursor].Name) {
		t.Fatalf("%d lines; last item not visible", len(lines))
	}
	for _, l := range lines {
		if displayWidth(l) > 50 {
			t.Fatalf("line too wide: %q", l)
		}
	}
	press(m, "home")
	if m.br.top != 0 {
		t.Fatalf("top = %d", m.br.top)
	}
}

func TestBrowseTopLevelHasNoDotDot(t *testing.T) {
	m, _, _ := browseModel(t)
	top := filepath.VolumeName(m.br.dir) + string(filepath.Separator)
	m.openBrowser(top, "")
	if len(m.br.items) > 0 && m.br.items[0].up {
		t.Fatal("file-system root must not offer ..")
	}
	press(m, "backspace") // must not panic or change anything
	if m.br.dir != top {
		t.Fatalf("dir = %q", m.br.dir)
	}
}
