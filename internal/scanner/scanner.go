// Package scanner walks a directory tree and computes how much space every
// folder uses, while keeping memory bounded: only folders that matter are
// retained, everything else is folded into per-folder "Other" buckets.
package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"space-explorer/internal/platform"
	"space-explorer/internal/tree"
)

// New prepares a scan of root.
func New(root string, opts Options) *Scanner {
	opts.normalize()
	s := &Scanner{
		root:  root,
		opts:  opts,
		skip:  make(map[string]struct{}, len(platform.SkipPaths)),
		stack: newTaskStack(),
		done:  make(chan struct{}),
		links: make(map[fileID]struct{}),
	}
	for _, p := range platform.SkipPaths {
		s.skip[p] = struct{}{}
	}
	return s
}

// Progress returns a snapshot of the counters; safe to call concurrently.
func (s *Scanner) Progress() Progress {
	p := Progress{
		Files:  s.files.Load(),
		Dirs:   s.dirs.Load(),
		Bytes:  s.bytes.Load(),
		Errors: s.errCount.Load(),
	}
	if c := s.current.Load(); c != nil {
		p.Current = *c
	}
	return p
}

// Run performs the scan. When ctx is cancelled it stops early and returns
// ctx.Err().
func (s *Scanner) Run(ctx context.Context) (*Result, error) {
	start := time.Now()

	abs, err := filepath.Abs(s.root)
	if err != nil {
		return nil, err
	}
	abs = filepath.Clean(abs)
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", abs)
	}
	u := platform.UsageOf(info, s.opts.Apparent)
	s.rootDev = u.Dev

	root := &dir{node: &tree.Node{Name: abs, IsDir: true}}
	if u.HasID && !s.opts.Apparent {
		root.self = u.Size
	}
	root.pending.Store(1)
	s.rootDir = root
	s.stack.push(root)

	var wg sync.WaitGroup
	for i := 0; i < s.opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx)
		}()
	}

	select {
	case <-s.done:
	case <-ctx.Done():
	}
	s.stack.close()
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.errMu.Lock()
	errs := append([]ScanError(nil), s.errs...)
	s.errMu.Unlock()

	return &Result{
		Root:       root.node,
		Errors:     errs,
		ErrorCount: s.errCount.Load(),
		Duration:   time.Since(start),
	}, nil
}

func (s *Scanner) worker(ctx context.Context) {
	for {
		d, ok := s.stack.pop()
		if !ok {
			return
		}
		if ctx.Err() != nil {
			return
		}
		s.process(ctx, d)
	}
}

func (s *Scanner) process(ctx context.Context, d *dir) {
	s.scanDir(ctx, d)
	if d.pending.Add(-1) == 0 {
		s.complete(d)
	}
}
