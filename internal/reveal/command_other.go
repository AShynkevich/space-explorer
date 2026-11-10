//go:build !darwin && !windows

package reveal

import "path/filepath"

// command: xdg-open can only open folders, so a file opens its parent folder.
func command(path string, isDir bool) (string, []string) {
	if !isDir {
		path = filepath.Dir(path)
	}
	return "xdg-open", []string{path}
}
