package tree

// Replace swaps a freshly scanned subtree in place of old and fixes the
// aggregated numbers (Size, Files, Dirs) and sort order of every ancestor.
// fresh takes over old's name and position; old must not be the root.
//
// Unreadable flags are only ever raised here, never cleared: an ancestor does
// not remember whether the problem came from this subtree or another one.
func Replace(old, fresh *Node) {
	p := old.Parent
	if p == nil {
		return
	}
	fresh.Name, fresh.Parent = old.Name, p
	for i, c := range p.Children {
		if c == old {
			p.Children[i] = fresh
			break
		}
	}
	dSize, dFiles, dDirs := fresh.Size-old.Size, fresh.Files-old.Files, fresh.Dirs-old.Dirs
	for a := p; a != nil; a = a.Parent {
		a.Size += dSize
		a.Files += dFiles
		a.Dirs += dDirs
		a.Unreadable = a.Unreadable || fresh.Unreadable
		SortBySize(a.Children)
	}
}
