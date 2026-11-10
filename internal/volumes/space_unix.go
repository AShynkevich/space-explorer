//go:build darwin || linux

package volumes

import "syscall"

// space returns total and user-available bytes of the file system at path.
func space(path string) (total, free int64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0
	}
	bs := uint64(st.Bsize)
	return int64(uint64(st.Blocks) * bs), int64(uint64(st.Bavail) * bs)
}
