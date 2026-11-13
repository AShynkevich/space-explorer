package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/session"
)

// eventMsg is a session event (progress, done, cancelled, failed) as a tea message.
type eventMsg struct {
	name string
	data any
}

// Bridge carries session events into the Bubble Tea loop. Pass Emit to
// session.New; the model re-arms wait() after every event it handles.
type Bridge struct{ ch chan eventMsg }

// NewBridge creates a Bridge with room for a burst of events.
func NewBridge() *Bridge { return &Bridge{ch: make(chan eventMsg, 64)} }

// Emit is a session.Manager emit function. Progress may be dropped when the
// UI is behind; terminal events (done, cancelled, failed) are never dropped.
func (b *Bridge) Emit(name string, data any) {
	e := eventMsg{name: name, data: data}
	if name == session.EventProgress {
		select {
		case b.ch <- e:
		default:
		}
		return
	}
	b.ch <- e
}

func (b *Bridge) wait() tea.Cmd {
	return func() tea.Msg { return <-b.ch }
}
