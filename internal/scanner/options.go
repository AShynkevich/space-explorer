package scanner

import "runtime"

// Options tunes the scan. The zero value is not useful; start from
// DefaultOptions.
type Options struct {
	// Workers is the number of folders listed concurrently. Kept low by
	// default so slow disks are not thrashed.
	Workers int
	// MinDirSize: finished folders smaller than this are folded into the
	// parent's Other bucket and freed.
	MinDirSize int64
	// MinFileSize: files smaller than this are never kept as nodes.
	MinFileSize int64
	// MaxFilesPerDir caps the number of file nodes kept per folder.
	MaxFilesPerDir int
	// Apparent reports file length instead of allocated disk blocks.
	// Always the case on platforms without block information (Windows).
	Apparent bool
	// OneFileSystem does not descend into folders on other devices.
	OneFileSystem bool
}

// DefaultOptions returns settings suited for low-memory machines.
func DefaultOptions() Options {
	w := runtime.NumCPU()
	if w < 2 {
		w = 2
	}
	if w > 4 {
		w = 4
	}
	return Options{
		Workers:        w,
		MinDirSize:     1 << 20,
		MinFileSize:    1 << 20,
		MaxFilesPerDir: 30,
	}
}

func (o *Options) normalize() {
	if o.Workers < 1 {
		o.Workers = 1
	}
	if o.MaxFilesPerDir < 0 {
		o.MaxFilesPerDir = 0
	}
	if o.MinDirSize < 0 {
		o.MinDirSize = 0
	}
	if o.MinFileSize < 0 {
		o.MinFileSize = 0
	}
}
