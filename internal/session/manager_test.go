package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAndBrowse(t *testing.T) {
	root := makeTree(t)
	rec := newRecorder()
	m := New(rec.emit, testOptions())

	if _, err := m.Children(root, ""); err == nil {
		t.Error("Children before any scan should fail")
	}
	if err := m.Start(root); err != nil {
		t.Fatal(err)
	}
	rec.wait(t, EventDone)

	sum := rec.last[EventDone].(Summary)
	if sum.Size != 4500 || sum.Files != 3 || sum.Dirs != 2 {
		t.Errorf("summary = %+v", sum)
	}

	l, err := m.Children(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if l.Total != 4500 || len(l.Entries) != 2 {
		t.Fatalf("listing = %+v", l)
	}
	first := l.Entries[0]
	if first.Name != "sub" || first.Kind != KindDir || !first.HasChildren || first.Percent < 77 || first.Percent > 78 {
		t.Errorf("first entry = %+v", first)
	}
	if l.Entries[1].Kind != KindFile || l.Entries[1].Name != "a.bin" {
		t.Errorf("second entry = %+v", l.Entries[1])
	}

	sub, err := m.Children(first.Path, "")
	if err != nil || len(sub.Entries) != 2 {
		t.Fatalf("sub listing = %+v, err %v", sub, err)
	}
	if _, err := m.Children(filepath.Join(root, "a.bin"), ""); err == nil {
		t.Error("Children of a file should fail")
	}
	if _, err := m.Children(filepath.Join(root, "nope"), ""); err == nil {
		t.Error("Children of an unknown path should fail")
	}
	if m.Running() {
		t.Error("Running() should be false after completion")
	}
}

func TestOtherRowAndCap(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < maxEntries+50; i++ {
		d := filepath.Join(root, "d"+string(rune('a'+i%26))+string(rune('a'+i/26%26))+string(rune('a'+i/676)))
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "f"), make([]byte, 10), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rec := newRecorder()
	m := New(rec.emit, testOptions())
	_ = m.Start(root)
	rec.wait(t, EventDone)

	l, err := m.Children(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Entries) != maxEntries+1 {
		t.Fatalf("entries = %d, want %d", len(l.Entries), maxEntries+1)
	}
	last := l.Entries[len(l.Entries)-1]
	if last.Kind != KindOther || last.Items != 50 || last.Size != 500 {
		t.Errorf("other row = %+v", last)
	}
}

func TestCancelAndRestart(t *testing.T) {
	root := makeTree(t)
	rec := newRecorder()
	m := New(rec.emit, testOptions())
	if err := m.Start(root); err != nil {
		t.Fatal(err)
	}
	m.Cancel() // may race with completion; both outcomes must leave a sane state
	if m.Running() {
		t.Error("Running() after Cancel")
	}
	if err := m.Start(root); err != nil { // restart works
		t.Fatal(err)
	}
	rec.wait(t, EventDone)
	if _, err := m.Children(root, ""); err != nil {
		t.Errorf("Children after restart: %v", err)
	}
}

func TestStartErrors(t *testing.T) {
	rec := newRecorder()
	m := New(rec.emit, testOptions())
	if err := m.Start(""); err == nil {
		t.Error("empty root should be rejected")
	}
	_ = m.Start(filepath.Join(t.TempDir(), "missing"))
	rec.wait(t, EventFailed)
	if m.Running() {
		t.Error("Running() after failure")
	}
}
