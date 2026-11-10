package tree

import (
	"path/filepath"
	"testing"
)

func TestFind(t *testing.T) {
	root := &Node{Name: filepath.FromSlash("/data"), IsDir: true}
	b := &Node{Name: "b", Parent: root, IsDir: true}
	c := &Node{Name: "c.bin", Parent: b}
	b.Children = []*Node{c}
	root.Children = []*Node{b}

	cases := []struct {
		path string
		want *Node
	}{
		{"/data", root},
		{"/data/", root},
		{"/data/b", b},
		{"/data/b/c.bin", c},
		{"/data/b/missing", nil},
		{"/data/../etc", nil},
		{"/other", nil},
	}
	for _, tc := range cases {
		if got := Find(root, filepath.FromSlash(tc.path)); got != tc.want {
			t.Errorf("Find(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
