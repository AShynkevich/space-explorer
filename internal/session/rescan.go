package session

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"space-explorer/internal/scanner"
	"space-explorer/internal/tree"
)

// Rescan re-reads one folder of the finished scan and swaps the fresh data
// into the existing tree (ancestor sizes are fixed up). The old data stays
// browsable until the rescan succeeds; cancelling or failing keeps it as is.
// Rescanning the scan root is the same as starting a new scan.
//
// Caveat: the fresh scan has its own hard-link set, so a file hard-linked
// from outside the folder is counted again inside it.
func (m *Manager) Rescan(path string) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return errors.New("a scan is already running")
	}
	base := m.result
	if base == nil {
		m.mu.Unlock()
		return errors.New("no finished scan")
	}
	n := tree.Find(base.Root, path)
	if n == nil || !n.IsDir {
		m.mu.Unlock()
		return fmt.Errorf("%s: not a scanned folder", path)
	}
	if n == base.Root {
		m.mu.Unlock()
		return m.Start(path)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	m.cancel, m.done, m.running = cancel, done, true
	m.mu.Unlock()

	go m.runRescan(ctx, base, n.Path(), done)
	return nil
}

func (m *Manager) runRescan(ctx context.Context, base *scanner.Result, path string, done chan struct{}) {
	defer close(done)
	res, err := m.scan(ctx, path)

	m.mu.Lock()
	m.running = false
	var sum Summary
	if err == nil {
		sum = m.graft(base, path, res)
	}
	m.mu.Unlock()
	m.finish(res, err, sum)
}

// graft replaces the folder at path inside base with the fresh result.
// Caller holds m.mu.
func (m *Manager) graft(base *scanner.Result, path string, res *scanner.Result) Summary {
	old := tree.Find(base.Root, path)
	if old == nil || m.result != base { // tree changed under us: drop the rescan
		return summarize(base, 0)
	}
	tree.Replace(old, res.Root)
	mergeErrors(base, path, res)
	s := summarize(base, res.Duration)
	s.Rescanned = path
	return s
}

// mergeErrors swaps the errors recorded below path for the fresh ones.
// ErrorCount is adjusted by the number of detailed entries that were removed,
// so it can be slightly off once more than the detail cap of errors exist.
func mergeErrors(base *scanner.Result, path string, fresh *scanner.Result) {
	kept := base.Errors[:0:0]
	for _, e := range base.Errors {
		if !within(e.Path, path) {
			kept = append(kept, e)
		}
	}
	removed := int64(len(base.Errors) - len(kept))
	base.Errors = append(kept, fresh.Errors...)
	base.ErrorCount += fresh.ErrorCount - removed
	if base.ErrorCount < int64(len(base.Errors)) {
		base.ErrorCount = int64(len(base.Errors))
	}
}

func within(p, dir string) bool {
	return p == dir || strings.HasPrefix(p, dir+string(filepath.Separator))
}
