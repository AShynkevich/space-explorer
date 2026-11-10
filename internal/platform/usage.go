// Package platform isolates OS-specific behaviour: how much disk space a file
// uses, file identity for hard-link detection, and paths never worth scanning.
package platform

// Usage is the platform-independent view of an lstat result.
type Usage struct {
	Size  int64  // bytes: allocated blocks, or apparent length
	Dev   uint64 // device id (valid if HasID)
	Ino   uint64 // inode (valid if HasID)
	Nlink uint64 // hard-link count (valid if HasID)
	HasID bool   // Dev/Ino/Nlink are valid
}
