package scanner

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"space-explorer/internal/platform"
	"space-explorer/internal/tree"
)

// scanDir lists one folder: files are aggregated locally, sub-folders become
// new tasks. The results are merged into n once, at the end.
func (s *Scanner) scanDir(ctx context.Context, d *dir) {
	n := d.node
	path := n.Path()
	s.current.Store(&path)

	var (
		files      int64
		total      int64 // sum of all file sizes in this folder
		other      tree.Other
		unreadable bool
		cands      []*tree.Node
	)
	max := s.opts.MaxFilesPerDir

	// trim keeps the `keep` largest candidates and folds the rest.
	trim := func(keep int) {
		tree.SortBySize(cands)
		for _, c := range cands[keep:] {
			other.Items++
			other.Size += c.Size
		}
		cands = cands[:keep]
	}

	f, err := os.Open(path)
	if err != nil {
		s.recordError(path, err)
		unreadable = true
	} else {
		defer f.Close()
		for ctx.Err() == nil {
			entries, err := f.ReadDir(readBatch)
			for _, e := range entries {
				name := e.Name()
				switch {
				case e.IsDir():
					s.queueDir(d, path, name)
				case e.Type().IsRegular():
					info, ierr := e.Info()
					if ierr != nil {
						s.recordError(filepath.Join(path, name), ierr)
						unreadable = true
						continue
					}
					u := platform.UsageOf(info, s.opts.Apparent)
					size := u.Size
					if u.HasID && u.Nlink > 1 && !s.firstLink(fileID{u.Dev, u.Ino}) {
						size = 0 // hard link already counted elsewhere
					}
					files++
					total += size
					s.files.Add(1)
					s.bytes.Add(size)
					if size < s.opts.MinFileSize || max == 0 {
						other.Items++
						other.Size += size
						continue
					}
					cands = append(cands, &tree.Node{Name: name, Parent: n, Size: size, Files: 1})
					if len(cands) >= 2*max {
						trim(max)
					}
				default:
					// symlinks, sockets, devices...: not followed, not counted
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					s.recordError(path, err)
					unreadable = true
				}
				break
			}
		}
	}
	if len(cands) > max {
		trim(max)
	}
	s.dirs.Add(1)

	// merge local results into the node
	d.mu.Lock()
	n.Files += files
	n.Size += total + d.self
	n.Other.Items += other.Items
	n.Other.Size += other.Size + d.self
	n.Unreadable = n.Unreadable || unreadable
	n.Children = append(n.Children, cands...)
	d.mu.Unlock()
}

// queueDir registers a sub-folder of parent as a new task.
func (s *Scanner) queueDir(parent *dir, parentPath, name string) {
	if len(s.skip) > 0 {
		if _, skip := s.skip[filepath.Join(parentPath, name)]; skip {
			return
		}
	}
	child := &dir{
		node:   &tree.Node{Name: name, Parent: parent.node, IsDir: true},
		parent: parent,
	}
	// Only folders with block information need an lstat here: it gives the
	// size of the folder entry itself (like `du`) and the device id.
	if s.opts.OneFileSystem || !s.opts.Apparent {
		if info, err := os.Lstat(filepath.Join(parentPath, name)); err == nil {
			u := platform.UsageOf(info, s.opts.Apparent)
			if u.HasID {
				if s.opts.OneFileSystem && u.Dev != s.rootDev {
					return
				}
				if !s.opts.Apparent {
					child.self = u.Size
				}
			}
		}
	}
	child.pending.Store(1)
	parent.pending.Add(1)
	s.stack.push(child)
}
