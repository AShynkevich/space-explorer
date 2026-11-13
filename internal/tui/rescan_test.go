package tui

import (
	"strings"
	"testing"

	"space-explorer/internal/session"
)

func TestRescanSubfolderKeepsTreeAndRefreshesOnDone(t *testing.T) {
	m, fb := newTree(t)
	press(m, "down", "right") // a expanded
	_, cmd := m.Update(keyRune('r'))
	cmd()
	if !m.busy || len(fb.rescans) != 1 || fb.rescans[0] != "/r/a" {
		t.Fatalf("busy=%v rescans=%v", m.busy, fb.rescans)
	}
	if m.mode != modeTree || len(m.rows) < 5 {
		t.Fatal("tree should stay browsable during a rescan")
	}
	fb.lists["/r/a"].Entries = fb.lists["/r/a"].Entries[:1]
	m.onEvent(eventMsg{session.EventDone, session.Summary{Root: "/r", Rescanned: "/r/a", DurationMs: 40}})
	if m.busy || !strings.Contains(m.status, "Rescanned /r/a") {
		t.Fatalf("busy=%v status=%q", m.busy, m.status)
	}
	if got, want := names(m), "/r,.a,..deep,.b,.f.bin,.3 other items"; got != want {
		t.Fatalf("rows = %q", got)
	}
}

func TestRescanRootIsFullScan(t *testing.T) {
	m, fb := newTree(t)
	_, cmd := m.Update(keyRune('r'))
	cmd()
	if m.mode != modeScanning || len(fb.started) == 0 {
		t.Fatalf("mode=%v started=%v", m.mode, fb.started)
	}
}

func TestRescanRefusedOnFileAndWhileBusy(t *testing.T) {
	m, fb := newTree(t)
	press(m, "down", "down", "down", "r") // f.bin
	if len(fb.rescans) != 0 || !strings.Contains(m.status, "folder") {
		t.Fatalf("status=%q", m.status)
	}
	m.busy = true
	press(m, "up", "up", "r")
	if len(fb.rescans) != 0 {
		t.Fatal("second rescan started")
	}
}

func TestEscCancelsBusyRescan(t *testing.T) {
	m, fb := newTree(t)
	m.busy = true
	_, cmd := m.Update(keyEsc())
	cmd()
	if fb.cancels != 1 {
		t.Fatalf("cancels = %d", fb.cancels)
	}
	m.onEvent(eventMsg{name: session.EventCancelled})
	if m.busy || m.mode != modeTree {
		t.Fatalf("busy=%v mode=%v", m.busy, m.mode)
	}
}

func TestErrorsPanelAndReveal(t *testing.T) {
	m, fb := newTree(t)
	fb.errs = []session.ErrorInfo{{Path: "/r/p", Err: "open /r/p: operation not permitted"}}
	m.summary.ErrorCount = 1
	var revealed string
	m.cfg.Reveal = func(p string) error { revealed = p; return nil }
	press(m, "e")
	if m.mode != modeErrors || !strings.Contains(m.View(), "Full Disk Access") {
		t.Fatalf("mode=%v", m.mode)
	}
	press(m, "enter")
	if revealed != "/r/p" {
		t.Fatalf("revealed %q", revealed)
	}
	press(m, "esc")
	press(m, "down", "o")
	if revealed != "/r/a" {
		t.Fatalf("revealed %q", revealed)
	}
}

func TestErrorsKeyWithNoErrors(t *testing.T) {
	m, _ := newTree(t)
	press(m, "e")
	if m.mode != modeTree || m.status != "No unreadable paths" {
		t.Fatalf("mode=%v status=%q", m.mode, m.status)
	}
}

func TestBackToDisksAndReturn(t *testing.T) {
	m, _ := newTree(t)
	press(m, "b")
	if m.mode != modeVolumes {
		t.Fatalf("mode=%v", m.mode)
	}
	press(m, "b")
	if m.mode != modeTree {
		t.Fatalf("mode=%v", m.mode)
	}
}

func TestQuitCancelsScan(t *testing.T) {
	m, fb := newTree(t)
	_, cmd := m.Update(keyRune('q'))
	if cmd == nil {
		t.Fatal("no command")
	}
	cmd()
	if fb.cancels != 1 {
		t.Fatalf("cancels = %d", fb.cancels)
	}
}
