package tui

import (
	"fmt"
	"strings"
	"testing"

	"space-explorer/internal/scanner"
	"space-explorer/internal/session"
)

func TestVolumePickStartsScan(t *testing.T) {
	fb := newFake()
	m := New(Config{Backend: fb, Bridge: NewBridge(), Home: "/home/u",
		Volumes: newTreeVolumes})
	_, cmd := m.Update(keyEnter())
	if m.mode != modeScanning || m.root != "/" {
		t.Fatalf("mode=%v root=%q", m.mode, m.root)
	}
	cmd()
	if len(fb.started) != 1 || fb.started[0] != "/" {
		t.Fatalf("started = %v", fb.started)
	}
}

func TestHomeShortcutAndPrompt(t *testing.T) {
	fb := newFake()
	m := New(Config{Backend: fb, Bridge: NewBridge(), Home: t.TempDir(), Volumes: newTreeVolumes})
	_, cmd := m.Update(keyRune('h'))
	cmd()
	if fb.started[0] != m.cfg.Home {
		t.Fatalf("started = %v", fb.started)
	}
	m.mode = modeVolumes
	press(m, "/", "~", "/", "x")
	press(m, "enter")
	if !strings.Contains(m.status, "not a folder") || m.mode != modePrompt {
		t.Fatalf("status=%q mode=%v", m.status, m.mode)
	}
	m.input = "~"
	press(m, "enter")
	if m.mode != modeScanning || m.root != m.cfg.Home {
		t.Fatalf("mode=%v root=%q", m.mode, m.root)
	}
}

func TestProgressAndDoneEvents(t *testing.T) {
	m, _ := newTree(t)
	m.beginScan("/r")
	m.onEvent(eventMsg{session.EventProgress, scanner.Progress{Files: 1234, Bytes: 2048, Current: "/r/a"}})
	if v := m.View(); !strings.Contains(v, "1,234 files") || !strings.Contains(v, "2.0 KB") {
		t.Fatalf("scanning view:\n%s", v)
	}
	m.onEvent(eventMsg{session.EventDone, session.Summary{Root: "/r", Size: 100}})
	if m.mode != modeTree || len(m.rows) == 0 {
		t.Fatalf("mode=%v rows=%d", m.mode, len(m.rows))
	}
}

func TestCancelledOrFailedFreshScanReturnsToDisks(t *testing.T) {
	m, _ := newTree(t)
	m.beginScan("/r")
	m.onEvent(eventMsg{name: session.EventFailed, data: "disk gone"})
	if m.mode != modeVolumes || !strings.Contains(m.status, "disk gone") {
		t.Fatalf("mode=%v status=%q", m.mode, m.status)
	}
	m.beginScan("/r")
	m.onEvent(eventMsg{name: session.EventCancelled})
	if m.mode != modeVolumes || m.status != "Scan cancelled" {
		t.Fatalf("mode=%v status=%q", m.mode, m.status)
	}
}

func TestStartFailureShownNotFatal(t *testing.T) {
	fb := newFake()
	fb.failing = true
	m := New(Config{Backend: fb, Bridge: NewBridge(), Volumes: newTreeVolumes})
	_, cmd := m.Update(keyEnter())
	m.Update(cmd())
	if m.mode != modeVolumes || !strings.Contains(m.status, "boom") {
		t.Fatalf("mode=%v status=%q", m.mode, m.status)
	}
}

func TestViewFitsTheTerminal(t *testing.T) {
	m, _ := newTree(t)
	for _, size := range [][2]int{{80, 24}, {40, 10}, {20, 5}, {200, 60}} {
		m.width, m.height = size[0], size[1]
		m.fixScroll()
		lines := strings.Split(m.View(), "\n")
		if len(lines) > size[1] {
			t.Fatalf("%v: %d lines", size, len(lines))
		}
		for _, l := range lines {
			if w := displayWidth(l); w > size[0] {
				t.Fatalf("%v: line width %d: %q", size, w, l)
			}
		}
	}
}

func TestBridgeNeverDropsTerminalEvents(t *testing.T) {
	b := NewBridge()
	for i := 0; i < 200; i++ {
		b.Emit(session.EventProgress, scanner.Progress{})
	}
	done := make(chan struct{})
	go func() { b.Emit(session.EventDone, session.Summary{}); close(done) }()
	got := ""
	for i := 0; i < 70 && got != session.EventDone; i++ {
		got = (<-b.ch).name
	}
	<-done
	if got != session.EventDone {
		t.Fatal(fmt.Sprint("done event lost, last=", got))
	}
}
