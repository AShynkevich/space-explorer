// Package tree defines the result model of a scan: a compact tree of folders
// (and the largest files) with aggregated sizes.
package tree

import (
	"path/filepath"
	"sort"
)

// Other aggregates entries that were folded away to keep memory bounded:
// small files, files beyond the per-folder cap and small sub-folders.
type Other struct {
	Items int64 // number of folded entries (files or folders)
	Size  int64 // total bytes of the folded entries
}

// Node is a retained entry of the scan tree: either a folder or one of the
// largest files of a folder.
//
// Invariant for folders: Size == sum(Children[i].Size) + Other.Size.
type Node struct {
	Name       string // base name; for the root node the full root path
	Parent     *Node
	Size       int64 // bytes used, including all descendants
	Files      int64 // regular files in the subtree (including folded ones)
	Dirs       int64 // folders in the subtree (excluding this one)
	IsDir      bool
	Unreadable bool    // folder (or part of it) could not be read
	Children   []*Node // retained children, sorted by Size descending
	Other      Other
}

// Path rebuilds the full path of the node from its parents.
func (n *Node) Path() string {
	depth := 0
	for p := n; p != nil; p = p.Parent {
		depth++
	}
	parts := make([]string, depth)
	for p := n; p != nil; p = p.Parent {
		depth--
		parts[depth] = p.Name
	}
	return filepath.Join(parts...)
}

// Count returns the number of retained nodes in the subtree, including n.
func (n *Node) Count() int {
	c := 1
	for _, ch := range n.Children {
		c += ch.Count()
	}
	return c
}

// SortBySize orders nodes by size descending, then by name.
func SortBySize(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Size != nodes[j].Size {
			return nodes[i].Size > nodes[j].Size
		}
		return nodes[i].Name < nodes[j].Name
	})
}
