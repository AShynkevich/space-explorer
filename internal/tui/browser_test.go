package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"space-explorer/internal/volumes"
)

// browseModel returns a model on the disk list whose only volume is a temp tree:
//
//	root/{alpha/{inner,leaf}, beta, gamma}, root/file.txt
func browseModel(t *testing.T) (*Model, *fakeBackend, string) {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"alpha/inner", "alpha/leaf", "beta", "gamma"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644)
	fb := newFake()
	m := New(Config{Backend: fb, Bridge: NewBridge(), Home: root,
		Volumes: func() []volumes.Volume { return []volumes.Volume{{Name: "T", Path: root}} }})
	return m, fb, root
}

func labels(m *Model) string {
	var l []string
	for _, it := range m.br.items {
		l = append(l, browseLabel(it))
	}
	return strings.Join(l, " ")
}

func TestBrowseFromDiskListShowsFoldersOnly(t *testing.T) {
	m, _, root := browseModel(t)
	press(m, "f")
	if m.mode != modeBrowse || m.br.dir != root {
		t.Fatalf("mode=%v dir=%q", m.mode, m.br.dir)
	}
	if got := labels(m); got != "../ alpha/ beta/ gamma/" {
		t.Fatalf("items = %q (files must not be listed)", got)
	}
}

func TestBrowseEnterDescendsAndBackspaceAscendsKeepingPlace(t *testing.T) {
	m, _, root := browseModel(t)
	press(m, "f", "down", "enter") // ../ -> alpha
	if m.br.dir != filepath.Join(root, "alpha") || labels(m) != "../ inner/ leaf/" {
		t.Fatalf("dir=%q items=%q", m.br.dir, labels(m))
	}
	press(m, "backspace")
	if m.br.dir != root {
		t.Fatalf("backspace should return to %q, got %q", root, m.br.dir)
	}
	press(m, "left")
	if want := filepath.Dir(root); m.br.dir != want {
		t.Fatalf("left should go above the root: %q, want %q", m.br.dir, want)
	}
}

func TestBrowseUpHighlightsFolderWeCameFrom(t *testing.T) {
	m, _, root := browseModel(t)
	m.openBrowser(filepath.Join(root, "gamma"), "")
	m.browseUp()
	if m.br.dir != root || m.br.items[m.br.cursor].Name != "gamma" {
		t.Fatalf("dir=%q cursor on %q", m.br.dir, m.br.items[m.br.cursor].Name)
	}
}

func TestBrowseDotDotEntryGoesUp(t *testing.T) {
	m, _, root := browseModel(t)
	m.openBrowser(filepath.Join(root, "beta"), "")
	press(m, "home", "enter")
	if m.br.dir != root {
		t.Fatalf("dir = %q", m.br.dir)
	}
}

func TestBrowseSpaceScansSelectedFolder(t *testing.T) {
	m, fb, root := browseModel(t)
	press(m, "f", "down", "down") // ../ -> alpha -> beta
	_, cmd := m.Update(keyRune(' '))
	cmd()
	want := filepath.Join(root, "beta")
	if m.mode != modeScanning || len(fb.started) != 1 || fb.started[0] != want {
		t.Fatalf("mode=%v started=%v want %s", m.mode, fb.started, want)
	}
}

func TestBrowseCScansCurrentFolder(t *testing.T) {
	m, fb, root := browseModel(t)
	press(m, "f")
	_, cmd := m.Update(keyRune('c'))
	cmd()
	if len(fb.started) != 1 || fb.started[0] != root {
		t.Fatalf("started = %v", fb.started)
	}
}

func TestBrowseScanOnDotDotExplainsInsteadOfScanning(t *testing.T) {
	m, fb, _ := browseModel(t)
	press(m, "f", " ")
	if len(fb.started) != 0 || m.mode != modeBrowse || !strings.Contains(m.status, "press c") {
		t.Fatalf("started=%v mode=%v status=%q", fb.started, m.mode, m.status)
	}
}

func TestBrowseCancelledScanReturnsToBrowser(t *testing.T) {
	m, _, root := browseModel(t)
	press(m, "f", "down", " ")
	m.onEvent(eventMsg{name: "scan:cancelled"})
	if m.mode != modeBrowse || m.br.dir != root {
		t.Fatalf("mode=%v dir=%q", m.mode, m.br.dir)
	}
}

func TestDiskListScanStillReturnsToDisksOnCancel(t *testing.T) {
	m, _, _ := browseModel(t)
	_, cmd := m.Update(keyEnter())
	cmd()
	m.onEvent(eventMsg{name: "scan:cancelled"})
	if m.mode != modeVolumes {
		t.Fatalf("mode=%v", m.mode)
	}
}
