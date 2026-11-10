package session

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"space-explorer/internal/scanner"
)

// recorder collects emitted events.
type recorder struct {
	mu     sync.Mutex
	events []string
	last   map[string]any
	ch     chan string
}

func newRecorder() *recorder {
	return &recorder{last: map[string]any{}, ch: make(chan string, 64)}
}

func (r *recorder) emit(name string, data any) {
	r.mu.Lock()
	r.events = append(r.events, name)
	r.last[name] = data
	r.mu.Unlock()
	select {
	case r.ch <- name:
	default:
	}
}

// wait blocks until the named event arrives.
func (r *recorder) wait(t *testing.T, name string) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case got := <-r.ch:
			if got == name {
				return
			}
		case <-timeout:
			t.Fatalf("timeout waiting for %s; got %v", name, r.events)
		}
	}
}

func testOptions() scanner.Options {
	o := scanner.DefaultOptions()
	o.Apparent, o.MinDirSize, o.MinFileSize = true, 0, 0
	return o
}

func makeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for p, size := range map[string]int{"a.bin": 1000, "sub/b.bin": 3000, "sub/deep/c.bin": 500} {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func scannerErr(path string) scanner.ScanError {
	return scanner.ScanError{Path: path, Err: "permission denied"}
}
