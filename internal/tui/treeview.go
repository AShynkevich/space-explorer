package tui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"space-explorer/internal/session"
)

const (
	sizeW = 10
	pctW  = 7
	barW  = 16
)

var permRe = regexp.MustCompile(`(?i)operation not permitted`)

const macHint = `macOS blocks some folders. Grant your terminal "Full Disk Access" in ` +
	`System Settings > Privacy & Security to include them, then rescan.`

func noteFor(items []session.ErrorInfo, total int64) string {
	var parts []string
	for _, e := range items {
		if permRe.MatchString(e.Err) {
			parts = append(parts, macHint)
			break
		}
	}
	if int64(len(items)) < total {
		parts = append(parts, fmt.Sprintf("Showing the first %s.", formatCount(int64(len(items)))))
	}
	return strings.Join(parts, " ")
}

func formatDurationMs(ms int64) string {
	return formatDuration(time.Duration(ms) * time.Millisecond)
}

func (m *Model) treeHead() string {
	arrow := map[string]string{session.SortSize: "size", session.SortName: "name"}[m.sortBy]
	return fmt.Sprintf("%s   [sorted by %s]", m.root, arrow)
}

func (m *Model) treeHelp() string {
	return "arrows move/open   s sort   r rescan   o reveal   e errors   b disks   q quit"
}

func (m *Model) treeView() string {
	bw := min(barW, max(0, m.width-sizeW-pctW-5-12)) // narrow terminals lose the bar first
	nameW := max(4, m.width-sizeW-pctW-bw-5)
	end := min(len(m.rows), m.top+m.bodyH())
	lines := make([]string, 0, end-m.top)
	for i := m.top; i < end; i++ {
		lines = append(lines, m.rowLine(m.rows[i], i == m.cursor, nameW, bw))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) rowLine(r row, selected bool, nameW, bw int) string {
	name := pad(trunc(indent(r)+marker(r)+r.e.Name, nameW), nameW)
	switch {
	case r.e.Kind == session.KindOther:
		name = otherStyle.Render(name)
	case r.e.Kind == session.KindDir:
		name = dirStyle.Render(name)
	case r.e.Unreadable:
		name = errStyle.Render(name)
	}
	line := fmt.Sprintf("%s %*s %*s %s", name, sizeW, formatSize(r.e.Size), pctW, formatPercent(r.e.Percent), renderBar(r.e.Percent, bw))
	line = trunc(line, m.width) // very narrow terminals cut the size columns
	if selected {
		line = selStyle.Render(line)
	}
	return line
}

func indent(r row) string { return strings.Repeat("  ", r.depth) }

func marker(r row) string {
	switch {
	case !r.expandable():
		return "  "
	case r.expanded:
		return "v "
	}
	return "> "
}

// pad right-pads s with spaces to w display cells.
func pad(s string, w int) string {
	if n := w - displayWidth(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}
