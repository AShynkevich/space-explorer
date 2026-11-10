package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"space-explorer/internal/tree"
)

// testOpts keeps everything (no folding) and counts apparent bytes so that
// expected sizes are exact on every platform.
func testOpts() Options {
	o := DefaultOptions()
	o.Apparent = true
	o.MinDirSize = 0
	o.MinFileSize = 0
	o.MaxFilesPerDir = 100
	return o
}

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, root string, o Options) *Result {
	t.Helper()
	res, err := New(root, o).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func find(n *tree.Node, name string) *tree.Node {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// checkInvariants verifies Size == sum(children) + Other and sort order.
func checkInvariants(t *testing.T, n *tree.Node) {
	t.Helper()
	var sum int64
	for i, c := range n.Children {
		sum += c.Size
		if i > 0 && n.Children[i-1].Size < c.Size {
			t.Errorf("%s: children not sorted by size", n.Path())
		}
		if c.IsDir {
			checkInvariants(t, c)
		}
	}
	if n.IsDir && sum+n.Other.Size != n.Size {
		t.Errorf("%s: size %d != children %d + other %d", n.Path(), n.Size, sum, n.Other.Size)
	}
}
