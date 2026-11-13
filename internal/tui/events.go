package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/scanner"
	"space-explorer/internal/session"
)

func (m *Model) onEvent(e eventMsg) tea.Cmd {
	switch e.name {
	case session.EventProgress:
		if p, ok := e.data.(scanner.Progress); ok {
			m.progress = p
			m.frame++
		}
	case session.EventDone:
		sum, _ := e.data.(session.Summary)
		m.onDone(sum)
	case session.EventCancelled:
		m.onStopped("Scan cancelled")
	case session.EventFailed:
		msg, _ := e.data.(string)
		m.onStopped("Scan failed: " + msg)
	}
	return m.cfg.Bridge.wait()
}

func (m *Model) onDone(sum session.Summary) {
	m.busy, m.summary, m.status = false, sum, ""
	if sum.Rescanned != "" {
		m.refresh()
		m.status = fmt.Sprintf("Rescanned %s in %s", sum.Rescanned,
			formatDuration(time.Duration(sum.DurationMs)*time.Millisecond))
		return
	}
	m.root = sum.Root
	m.load()
}

// onStopped handles a cancelled or failed scan. A failed fresh scan returns
// to the disk list; a stopped rescan leaves the old tree as it was.
func (m *Model) onStopped(msg string) {
	m.status, m.busy = msg, false
	if m.mode == modeScanning {
		m.mode = m.cancelTo
	}
}
