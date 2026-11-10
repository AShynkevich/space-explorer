package scanner

import (
	"sync"
	"sync/atomic"
	"time"

	"space-explorer/internal/tree"
)

const (
	readBatch = 1024 // directory entries read per call
	maxErrors = 500  // detailed errors kept; the rest is only counted
)

type fileID struct{ dev, ino uint64 }

// ScanError describes a path that could not be read.
type ScanError struct {
	Path string
	Err  string
}

// Progress is a cheap snapshot of a running scan.
type Progress struct {
	Files   int64
	Dirs    int64
	Bytes   int64
	Errors  int64
	Current string // folder being listed (approximate)
}

// Result is returned by a finished scan.
type Result struct {
	Root       *tree.Node
	Errors     []ScanError // at most maxErrors entries
	ErrorCount int64       // total number of errors, may exceed len(Errors)
	Duration   time.Duration
}

// Scanner scans one root folder. Create it with New, call Run once.
type Scanner struct {
	root string
	opts Options
	skip map[string]struct{}

	files, dirs, bytes, errCount atomic.Int64
	current                      atomic.Pointer[string]

	rootDev uint64
	rootDir *dir

	stack *taskStack
	done  chan struct{}
	once  sync.Once

	errMu sync.Mutex
	errs  []ScanError

	linkMu sync.Mutex
	links  map[fileID]struct{}
}

// dir is the scan-time state of a folder that is queued or being listed. It
// is dropped when the folder completes; only the tree.Node stays.
type dir struct {
	node    *tree.Node
	parent  *dir
	mu      sync.Mutex   // guards node's aggregated fields while merging
	pending atomic.Int32 // 1 (own listing) + number of unfinished sub-folders
	self    int64        // size of the folder entry itself (allocated blocks)
}
