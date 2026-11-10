//go:build darwin || linux || freebsd || netbsd || openbsd

package platform

import (
	"io/fs"
	"syscall"
)

// UsageOf extracts size and identity information from a Lstat result.
func UsageOf(info fs.FileInfo, apparent bool) Usage {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return Usage{Size: info.Size()}
	}
	u := Usage{
		Size:  int64(st.Blocks) * 512, // allocated space, like `du`
		Dev:   uint64(st.Dev),
		Ino:   uint64(st.Ino),
		Nlink: uint64(st.Nlink),
		HasID: true,
	}
	if apparent {
		u.Size = info.Size()
	}
	return u
}
