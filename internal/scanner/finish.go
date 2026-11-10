package scanner

import "space-explorer/internal/tree"

// complete finishes d and every ancestor whose last pending child it was.
func (s *Scanner) complete(d *dir) {
	for d != nil {
		n := d.node
		tree.SortBySize(n.Children)
		p := d.parent
		if p == nil {
			s.once.Do(func() { close(s.done) })
			return
		}
		pn := p.node
		p.mu.Lock()
		pn.Size += n.Size
		pn.Files += n.Files
		pn.Dirs += n.Dirs + 1
		if n.Size >= s.opts.MinDirSize {
			pn.Children = append(pn.Children, n)
		} else {
			pn.Other.Items++
			pn.Other.Size += n.Size
			n.Children = nil
		}
		pn.Unreadable = pn.Unreadable || n.Unreadable
		p.mu.Unlock()
		if p.pending.Add(-1) != 0 {
			return
		}
		d = p
	}
}

func (s *Scanner) firstLink(id fileID) bool {
	s.linkMu.Lock()
	defer s.linkMu.Unlock()
	if _, seen := s.links[id]; seen {
		return false
	}
	s.links[id] = struct{}{}
	return true
}

func (s *Scanner) recordError(path string, err error) {
	s.errCount.Add(1)
	s.errMu.Lock()
	if len(s.errs) < maxErrors {
		s.errs = append(s.errs, ScanError{Path: path, Err: err.Error()})
	}
	s.errMu.Unlock()
}
