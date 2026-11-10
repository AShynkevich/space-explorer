package volumes

import (
	"os"
	"path/filepath"
)

// list returns "/" plus everything mounted under /Volumes. The boot volume
// shows up in /Volumes as a symlink to "/": it only provides the label.
func list() []Volume {
	root := Volume{Name: "Macintosh HD", Path: "/"}
	var out []Volume
	entries, _ := os.ReadDir("/Volumes")
	for _, e := range entries {
		p := filepath.Join("/Volumes", e.Name())
		if e.Type()&os.ModeSymlink != 0 {
			if target, err := filepath.EvalSymlinks(p); err == nil && target == "/" {
				root.Name = e.Name()
			}
			continue
		}
		if e.IsDir() {
			out = append(out, Volume{Name: e.Name(), Path: p})
		}
	}
	return append([]Volume{root}, out...)
}
