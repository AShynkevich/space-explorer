package tree

import (
	"path/filepath"
	"strings"
)

// Find returns the retained node for an absolute path inside the tree rooted
// at root, or nil if the path is outside the root or was folded away.
func Find(root *Node, path string) *Node {
	rel, err := filepath.Rel(root.Name, filepath.Clean(path))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil
	}
	if rel == "." {
		return root
	}
	n := root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		n = child(n, part)
		if n == nil {
			return nil
		}
	}
	return n
}

func child(n *Node, name string) *Node {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}
