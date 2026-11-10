// Package volumes lists the disks/volumes a user can pick as a scan root.
package volumes

import "sort"

// Volume is a mounted disk or network share.
type Volume struct {
	Name  string `json:"name"`  // human-friendly label
	Path  string `json:"path"`  // mount point / drive root
	Total int64  `json:"total"` // capacity in bytes (0 if unknown)
	Free  int64  `json:"free"`  // bytes available to the user (0 if unknown)
}

// List returns the volumes of this machine, the system volume first.
func List() []Volume {
	vs := list()
	for i := range vs {
		vs[i].Total, vs[i].Free = space(vs[i].Path)
	}
	sort.SliceStable(vs, func(i, j int) bool { return isSystem(vs[i].Path) && !isSystem(vs[j].Path) })
	return vs
}

func isSystem(path string) bool {
	return path == "/" || path == `C:\`
}
