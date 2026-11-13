// Package browse lists the sub-folders of a directory so a UI can let the
// user pick a scan root by walking the file system (file-manager style).
// It does no scanning and knows nothing about any UI.
package browse

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is a sub-folder of the listed directory.
type Entry struct {
	Name string
	Path string
	Link bool // a symbolic link that points to a folder
}

// List returns the sub-folders of dir, sorted by name (case-insensitive).
// Files are left out, symlinks to folders are included (marked Link) and
// broken symlinks are skipped. A directory that cannot be read at all is an
// error; one that is partly readable returns what could be listed.
func List(dir string) ([]Entry, error) {
	des, err := os.ReadDir(dir)
	if err != nil && len(des) == 0 {
		return nil, err
	}
	out := make([]Entry, 0, len(des))
	for _, de := range des {
		p := filepath.Join(dir, de.Name())
		switch {
		case de.IsDir():
			out = append(out, Entry{Name: de.Name(), Path: p})
		case de.Type()&os.ModeSymlink != 0:
			if info, serr := os.Stat(p); serr == nil && info.IsDir() {
				out = append(out, Entry{Name: de.Name(), Path: p, Link: true})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i].Name), strings.ToLower(out[j].Name)
		if a != b {
			return a < b
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// Parent returns the folder above dir. ok is false at a file-system root
// ("/", "C:\"), where there is nothing above.
func Parent(dir string) (parent string, ok bool) {
	dir = filepath.Clean(dir)
	p := filepath.Dir(dir)
	return p, p != dir
}
