package tui

import "space-explorer/internal/session"

// row is one visible line of the flat tree.
type row struct {
	e        session.Entry
	depth    int
	expanded bool
}

// The "other" row shares its parent's path, so identity is kind + path.
func entryKey(e session.Entry) string { return e.Kind + "|" + e.Path }

func (r row) key() string { return entryKey(r.e) }

func (r row) expandable() bool { return r.e.Kind == session.KindDir && r.e.HasChildren }

type fetchFn func(path string) (*session.Listing, error)

func rowsOf(l *session.Listing, depth int) []row {
	out := make([]row, len(l.Entries))
	for i, e := range l.Entries {
		out[i] = row{e: e, depth: depth}
	}
	return out
}

// buildRows lists l and recurses into the folders in open (restores expanded
// state after a refresh). Folders that fail to load stay collapsed.
func buildRows(fetch fetchFn, l *session.Listing, depth int, open map[string]bool) []row {
	var out []row
	for _, e := range l.Entries {
		r := row{e: e, depth: depth}
		out = append(out, r)
		if !r.expandable() || !open[e.Path] {
			continue
		}
		nested, err := fetch(e.Path)
		if err != nil {
			continue
		}
		out[len(out)-1].expanded = true
		out = append(out, buildRows(fetch, nested, depth+1, open)...)
	}
	return out
}

// loadTree returns the scan root (expanded, 100%) followed by its content.
func loadTree(fetch fetchFn, root string, open map[string]bool) ([]row, error) {
	l, err := fetch(root)
	if err != nil {
		return nil, err
	}
	top := row{expanded: true, e: session.Entry{
		Name: root, Path: root, Kind: session.KindDir, Size: l.Size,
		Percent: 100, HasChildren: len(l.Entries) > 0,
	}}
	return append([]row{top}, buildRows(fetch, l, 1, open)...), nil
}

func insertAfter(rows []row, idx int, kids []row) []row {
	out := make([]row, 0, len(rows)+len(kids))
	out = append(out, rows[:idx+1]...)
	out = append(out, kids...)
	return append(out, rows[idx+1:]...)
}

// collapseAt hides everything below rows[idx]; nested expansions are dropped.
func collapseAt(rows []row, idx int) []row {
	end := idx + 1
	for end < len(rows) && rows[end].depth > rows[idx].depth {
		end++
	}
	rows[idx].expanded = false
	return append(rows[:idx+1], rows[end:]...)
}

func openSet(rows []row) map[string]bool {
	open := map[string]bool{}
	for _, r := range rows {
		if r.expanded && r.e.Kind == session.KindDir {
			open[r.e.Path] = true
		}
	}
	return open
}

func indexOfKey(rows []row, key string) int {
	for i, r := range rows {
		if r.key() == key {
			return i
		}
	}
	return -1
}

// parentIndex is the nearest row above idx with a smaller depth, or -1.
func parentIndex(rows []row, idx int) int {
	for i := idx - 1; i >= 0; i-- {
		if rows[i].depth < rows[idx].depth {
			return i
		}
	}
	return -1
}
