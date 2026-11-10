// Package session owns the lifecycle of the current scan for the UI: start,
// cancel, throttled progress events and lazy access to the finished tree.
// It knows nothing about Wails; events go through an injected Emit function.
package session

import (
	"context"
	"errors"
	"sync"

	"space-explorer/internal/scanner"
)

// Manager runs at most one scan at a time and keeps its result in memory.
type Manager struct {
	emit func(event string, data any)
	opts scanner.Options

	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{} // closed when the active run goroutine ends
	result  *scanner.Result
	running bool
}

// New creates a Manager. emit may be nil (events are dropped).
func New(emit func(event string, data any), opts scanner.Options) *Manager {
	if emit == nil {
		emit = func(string, any) {}
	}
	return &Manager{emit: emit, opts: opts}
}

// Start begins scanning root. A scan already in progress is cancelled first
// and the previous result is released to keep memory low.
func (m *Manager) Start(root string) error {
	if root == "" {
		return errors.New("no folder selected")
	}
	m.Cancel()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	m.mu.Lock()
	m.cancel, m.done, m.result, m.running = cancel, done, nil, true
	m.mu.Unlock()

	go m.run(ctx, root, done)
	return nil
}

// Cancel stops the running scan, if any, and waits until it has ended.
func (m *Manager) Cancel() {
	m.mu.Lock()
	cancel, done := m.cancel, m.done
	m.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
}

// Running reports whether a scan is in progress.
func (m *Manager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// Errors returns the paths that could not be read in the last scan.
func (m *Manager) Errors() []ErrorInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := m.result
	if res == nil {
		return nil
	}
	out := make([]ErrorInfo, len(res.Errors))
	for i, e := range res.Errors {
		out[i] = ErrorInfo{Path: e.Path, Err: e.Err}
	}
	return out
}

func (m *Manager) run(ctx context.Context, root string, done chan struct{}) {
	defer close(done)
	res, err := m.scan(ctx, root)

	m.mu.Lock()
	m.running = false
	if err == nil {
		m.result = res
	}
	m.mu.Unlock()
	m.finish(res, err, Summary{})
}

// scan runs one scanner with throttled progress events.
func (m *Manager) scan(ctx context.Context, root string) (*scanner.Result, error) {
	s := scanner.New(root, m.opts)
	stop := make(chan struct{})
	pumped := make(chan struct{})
	go func() {
		defer close(pumped)
		pump(s, m.emit, stop)
	}()
	res, err := s.Run(ctx)
	close(stop)
	<-pumped
	return res, err
}

// finish emits the terminal event. sum is used for rescans; for a fresh scan
// (zero Summary) it is derived from res.
func (m *Manager) finish(res *scanner.Result, err error, sum Summary) {
	switch {
	case errors.Is(err, context.Canceled):
		m.emit(EventCancelled, nil)
	case err != nil:
		m.emit(EventFailed, err.Error())
	case sum.Root != "":
		m.emit(EventDone, sum)
	default:
		m.emit(EventDone, summarize(res, res.Duration))
	}
}
