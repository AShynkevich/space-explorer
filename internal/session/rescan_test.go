package session

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSized(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func scanned(t *testing.T, root string) (*Manager, *recorder) {
	t.Helper()
	rec := newRecorder()
	m := New(rec.emit, testOptions())
	if err := m.Start(root); err != nil {
		t.Fatal(err)
	}
	rec.wait(t, EventDone)
	return m, rec
}

func TestRescanSubfolderUpdatesTotals(t *testing.T) {
	root := makeTree(t) // a.bin 1000, sub/b.bin 3000, sub/deep/c.bin 500 => 4500
	m, rec := scanned(t, root)

	writeSized(t, filepath.Join(root, "sub", "deep", "new.bin"), 5000)
	if err := os.Remove(filepath.Join(root, "sub", "b.bin")); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub")
	if err := m.Rescan(sub); err != nil {
		t.Fatal(err)
	}
	rec.wait(t, EventDone)

	sum := rec.last[EventDone].(Summary)
	// 1000 (a) + 500 (c) + 5000 (new) = 6500
	if sum.Size != 6500 || sum.Rescanned != sub || sum.Root != root {
		t.Errorf("summary = %+v", sum)
	}
	l, err := m.Children(root, "")
	if err != nil || l.Total != 6500 || l.Entries[0].Name != "sub" || l.Entries[0].Size != 5500 {
		t.Fatalf("root listing = %+v err %v", l, err)
	}
	deep, err := m.Children(filepath.Join(sub, "deep"), "")
	if err != nil || len(deep.Entries) != 2 || deep.Entries[0].Name != "new.bin" {
		t.Errorf("deep listing = %+v err %v", deep, err)
	}
	// percentages follow the new total
	if p := l.Entries[0].Percent; p < 84 || p > 85 {
		t.Errorf("percent = %f, want ~84.6", p)
	}
}

func TestRescanRootEqualsFullScan(t *testing.T) {
	root := makeTree(t)
	m, rec := scanned(t, root)
	writeSized(t, filepath.Join(root, "extra"), 700)
	if err := m.Rescan(root); err != nil {
		t.Fatal(err)
	}
	rec.wait(t, EventDone)
	if sum := rec.last[EventDone].(Summary); sum.Size != 5200 || sum.Rescanned != "" {
		t.Errorf("summary = %+v", sum)
	}
}

func TestRescanRejectsBadInput(t *testing.T) {
	root := makeTree(t)
	rec := newRecorder()
	m := New(rec.emit, testOptions())
	if err := m.Rescan(root); err == nil {
		t.Error("rescan before any scan should fail")
	}
	m, _ = scanned(t, root)
	if err := m.Rescan(filepath.Join(root, "nope")); err == nil {
		t.Error("unknown path should fail")
	}
	if err := m.Rescan(filepath.Join(root, "a.bin")); err == nil {
		t.Error("a file should fail")
	}
}

func TestRescanCancelKeepsOldData(t *testing.T) {
	root := makeTree(t)
	m, rec := scanned(t, root)
	if err := m.Rescan(filepath.Join(root, "sub")); err != nil {
		t.Fatal(err)
	}
	m.Cancel() // may race with completion; either way the tree must stay consistent
	l, err := m.Children(root, "")
	if err != nil || l.Total != 4500 {
		t.Errorf("after cancelled rescan: %+v err %v", l, err)
	}
	if m.Running() {
		t.Error("still running after Cancel")
	}
	_ = rec
}

func TestRescanErrorsAreMerged(t *testing.T) {
	root := makeTree(t)
	m, rec := scanned(t, root)
	m.mu.Lock() // pretend the first scan saw errors inside and outside sub
	m.result.Errors = append(m.result.Errors,
		scannerErr(filepath.Join(root, "sub", "old")), scannerErr(filepath.Join(root, "other")))
	m.result.ErrorCount = 2
	m.mu.Unlock()

	if err := m.Rescan(filepath.Join(root, "sub")); err != nil {
		t.Fatal(err)
	}
	rec.wait(t, EventDone)
	errs := m.Errors()
	if len(errs) != 1 || errs[0].Path != filepath.Join(root, "other") {
		t.Errorf("errors = %+v, want only the one outside sub", errs)
	}
}

func TestChildrenSortByName(t *testing.T) {
	root := t.TempDir()
	writeSized(t, filepath.Join(root, "bravo", "f"), 3000)
	writeSized(t, filepath.Join(root, "Alpha", "f"), 1000)
	writeSized(t, filepath.Join(root, "charlie", "f"), 2000)
	m, _ := scanned(t, root)

	names := func(sortBy string) []string {
		l, err := m.Children(root, sortBy)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, e := range l.Entries {
			out = append(out, e.Name)
		}
		return out
	}
	eq := func(got, want []string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	}
	eq(names(""), []string{"bravo", "charlie", "Alpha"})
	eq(names(SortSize), []string{"bravo", "charlie", "Alpha"})
	eq(names(SortName), []string{"Alpha", "bravo", "charlie"})
}
