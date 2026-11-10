package tree

import (
	"path/filepath"
	"testing"
)

func TestPathCountSort(t *testing.T) {
	root := &Node{Name: filepath.FromSlash("/data"), IsDir: true}
	a := &Node{Name: "a", Parent: root, Size: 10, IsDir: true}
	b := &Node{Name: "b", Parent: root, Size: 30, IsDir: true}
	c := &Node{Name: "c.bin", Parent: b, Size: 5}
	b.Children = []*Node{c}
	root.Children = []*Node{a, b}

	if got, want := c.Path(), filepath.FromSlash("/data/b/c.bin"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
	if got := root.Count(); got != 4 {
		t.Errorf("Count() = %d, want 4", got)
	}
	SortBySize(root.Children)
	if root.Children[0] != b {
		t.Error("SortBySize: largest should be first")
	}
}
