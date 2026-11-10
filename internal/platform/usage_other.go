//go:build !(darwin || linux || freebsd || netbsd || openbsd)

package platform

import "io/fs"

// UsageOf falls back to the apparent size where block information or inode
// identity is not available (e.g. Windows).
func UsageOf(info fs.FileInfo, _ bool) Usage {
	return Usage{Size: info.Size()}
}
