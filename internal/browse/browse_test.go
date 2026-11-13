package browse

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func mk(t *testing.T, root string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func names(es []Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Name
	}
	return out
}

func TestListSortsFoldersOnlyIncludingHidden(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "beta", "Alpha", ".hidden", "alpha2")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	es, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	got := names(es)
	want := []string{".hidden", "Alpha", "alpha2", "beta"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if es[1].Path != filepath.Join(root, "Alpha") {
		t.Fatalf("path = %q", es[1].Path)
	}
}

func TestListEmptyDirIsNotAnError(t *testing.T) {
	es, err := List(t.TempDir())
	if err != nil || len(es) != 0 {
		t.Fatalf("es=%v err=%v", es, err)
	}
}

func TestListMissingDirFails(t *testing.T) {
	if _, err := List(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected an error")
	}
}

func TestListSymlinks(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "real")
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks not available:", err)
	}
	_ = os.Symlink(filepath.Join(root, "gone"), filepath.Join(root, "broken"))
	_ = os.WriteFile(filepath.Join(root, "f"), nil, 0o644)
	_ = os.Symlink(filepath.Join(root, "f"), filepath.Join(root, "filelink"))

	es, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(es); len(got) != 2 || got[0] != "link" || got[1] != "real" {
		t.Fatalf("got %v, want [link real]", got)
	}
	if !es[0].Link || es[1].Link {
		t.Fatalf("link flags: %+v", es)
	}
}

func TestListUnreadableDirFails(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission bits are not enforced here")
	}
	root := t.TempDir()
	mk(t, root, "locked")
	p := filepath.Join(root, "locked")
	if err := os.Chmod(p, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(p, 0o755)
	if _, err := List(p); err == nil {
		t.Fatal("expected a permission error")
	}
}

func TestParent(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a", "b")
	if p, ok := Parent(child); !ok || p != filepath.Join(root, "a") {
		t.Fatalf("Parent(%q) = %q, %v", child, p, ok)
	}
	if p, ok := Parent(child + string(filepath.Separator)); !ok || p != filepath.Join(root, "a") {
		t.Fatalf("trailing separator: %q, %v", p, ok)
	}
	top := filepath.VolumeName(root) + string(filepath.Separator)
	if _, ok := Parent(top); ok {
		t.Fatalf("%q should have no parent", top)
	}
}
