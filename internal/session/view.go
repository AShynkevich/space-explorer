package session

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"space-explorer/internal/tree"
)

// maxEntries caps the rows returned for one folder so the DOM stays small;
// the remainder is merged into the "other" row.
const maxEntries = 200

// Entry kinds.
const (
	KindDir   = "dir"
	KindFile  = "file"
	KindOther = "other" // synthetic row: folded small files/folders
)

// Entry is one row of the tree view.
type Entry struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Kind        string  `json:"kind"`
	Size        int64   `json:"size"`
	Percent     float64 `json:"percent"` // share of the scan root, 0..100
	Items       int64   `json:"items"`   // files (dir) or folded entries (other)
	HasChildren bool    `json:"hasChildren"`
	Unreadable  bool    `json:"unreadable"`
}

// Listing is the content of one folder.
type Listing struct {
	Path    string  `json:"path"`
	Size    int64   `json:"size"`
	Total   int64   `json:"total"` // size of the scan root
	Entries []Entry `json:"entries"`
}

// Sort orders for Children.
const (
	SortSize = "size" // largest first (default)
	SortName = "name" // A-Z, case-insensitive
)

// Children returns the entries of the folder at path: the 200 largest, in the
// requested order (SortSize if empty or unknown), then the synthetic "other"
// row. The lock is held while building so a concurrent rescan cannot swap
// nodes mid-read.
func (m *Manager) Children(path, sortBy string) (*Listing, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := m.result
	if res == nil {
		return nil, errors.New("no finished scan")
	}
	n := tree.Find(res.Root, path)
	if n == nil {
		return nil, fmt.Errorf("%s: not in scan result", path)
	}
	if !n.IsDir {
		return nil, fmt.Errorf("%s is not a folder", path)
	}
	l := buildListing(n, res.Root.Size)
	sortEntries(l.Entries, sortBy)
	return l, nil
}

func buildListing(n *tree.Node, total int64) *Listing {
	l := &Listing{Path: n.Path(), Size: n.Size, Total: total}
	other := n.Other
	for i, c := range n.Children {
		if i >= maxEntries {
			other.Items++
			other.Size += c.Size
			continue
		}
		kind := KindFile
		if c.IsDir {
			kind = KindDir
		}
		l.Entries = append(l.Entries, Entry{
			Name:        c.Name,
			Path:        c.Path(),
			Kind:        kind,
			Size:        c.Size,
			Percent:     percent(c.Size, total),
			Items:       c.Files,
			HasChildren: len(c.Children) > 0,
			Unreadable:  c.Unreadable,
		})
	}
	if other.Items > 0 {
		l.Entries = append(l.Entries, Entry{
			Name:    fmt.Sprintf("%d other items", other.Items),
			Path:    n.Path(),
			Kind:    KindOther,
			Size:    other.Size,
			Percent: percent(other.Size, total),
			Items:   other.Items,
		})
	}
	return l
}

func percent(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

// sortEntries reorders rows for the requested mode; the "other" row stays last.
func sortEntries(entries []Entry, sortBy string) {
	if sortBy != SortName {
		return // buildListing already returns largest first
	}
	if n := len(entries); n > 0 && entries[n-1].Kind == KindOther {
		entries = entries[:n-1]
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}
