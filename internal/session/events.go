package session

import (
	"time"

	"space-explorer/internal/scanner"
)

// Event names sent through Options.Emit.
const (
	EventProgress  = "scan:progress"  // payload: scanner.Progress (throttled)
	EventDone      = "scan:done"      // payload: Summary
	EventCancelled = "scan:cancelled" // payload: nil
	EventFailed    = "scan:failed"    // payload: string (message)
)

// progressInterval throttles progress events (about 4 per second).
const progressInterval = 250 * time.Millisecond

// Summary describes a finished scan.
type Summary struct {
	Root       string `json:"root"`
	Size       int64  `json:"size"`
	Files      int64  `json:"files"`
	Dirs       int64  `json:"dirs"`
	ErrorCount int64  `json:"errorCount"`
	DurationMs int64  `json:"durationMs"`
	Rescanned  string `json:"rescanned,omitempty"` // set when only a sub-folder was rescanned
}

// ErrorInfo mirrors scanner.ScanError for the UI.
type ErrorInfo struct {
	Path string `json:"path"`
	Err  string `json:"err"`
}

// summarize describes the tree held in r; dur is the time of the scan that produced it.
func summarize(r *scanner.Result, dur time.Duration) Summary {
	return Summary{
		Root:       r.Root.Name,
		Size:       r.Root.Size,
		Files:      r.Root.Files,
		Dirs:       r.Root.Dirs,
		ErrorCount: r.ErrorCount,
		DurationMs: dur.Milliseconds(),
	}
}

// pump emits progress snapshots until stop is closed.
func pump(s *scanner.Scanner, emit func(string, any), stop <-chan struct{}) {
	t := time.NewTicker(progressInterval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			emit(EventProgress, s.Progress())
		}
	}
}
