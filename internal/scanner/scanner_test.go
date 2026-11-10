package scanner

import (
	"path/filepath"
	"testing"
)

func TestSizesAggregate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.bin"), 1000)
	writeFile(t, filepath.Join(root, "sub", "b.bin"), 2000)
	writeFile(t, filepath.Join(root, "sub", "deep", "c.bin"), 4000)
	writeFile(t, filepath.Join(root, "other", "d.bin"), 500)

	res := run(t, root, testOpts())
	r := res.Root
	if r.Size != 7500 {
		t.Errorf("root size = %d, want 7500", r.Size)
	}
	if r.Files != 4 || r.Dirs != 3 {
		t.Errorf("root files/dirs = %d/%d, want 4/3", r.Files, r.Dirs)
	}
	sub := find(r, "sub")
	if sub == nil || sub.Size != 6000 {
		t.Fatalf("sub = %+v, want size 6000", sub)
	}
	if deep := find(sub, "deep"); deep == nil || deep.Size != 4000 {
		t.Errorf("deep = %+v, want size 4000", deep)
	}
	if got := find(sub, "deep").Path(); got != filepath.Join(root, "sub", "deep") {
		t.Errorf("Path() = %q", got)
	}
	// biggest folder first
	if r.Children[0].Name != "sub" {
		t.Errorf("first child = %q, want sub", r.Children[0].Name)
	}
	checkInvariants(t, r)
}

func TestFoldingKeepsTotals(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 10; i++ {
		writeFile(t, filepath.Join(root, "f"+string(rune('a'+i))), 100*(i+1))
	}
	writeFile(t, filepath.Join(root, "tiny", "x"), 10)
	writeFile(t, filepath.Join(root, "big", "y"), 5000)

	o := testOpts()
	o.MaxFilesPerDir = 3
	o.MinDirSize = 1000
	res := run(t, root, o)
	r := res.Root

	if want := int64(100*55 + 10 + 5000); r.Size != want {
		t.Errorf("root size = %d, want %d", r.Size, want)
	}
	if find(r, "tiny") != nil {
		t.Error("tiny folder should have been folded away")
	}
	if find(r, "big") == nil {
		t.Error("big folder should be retained")
	}
	files := 0
	for _, c := range r.Children {
		if !c.IsDir {
			files++
		}
	}
	if files != 3 {
		t.Errorf("kept %d file nodes, want 3", files)
	}
	// 7 files + 1 tiny dir folded
	if r.Other.Items != 8 {
		t.Errorf("other items = %d, want 8", r.Other.Items)
	}
	// the three largest files are kept
	for _, n := range []string{"fj", "fi", "fh"} {
		if find(r, n) == nil {
			t.Errorf("expected large file %s to be kept", n)
		}
	}
	checkInvariants(t, r)
}
func TestEmptyRoot(t *testing.T) {
	res := run(t, t.TempDir(), testOpts())
	if res.Root.Size != 0 || len(res.Root.Children) != 0 {
		t.Errorf("empty root: %+v", res.Root)
	}
}
