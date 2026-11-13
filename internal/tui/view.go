package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (m *Model) View() string {
	var head, body, foot string
	switch m.mode {
	case modeVolumes:
		head, body, foot = "Space Explorer - choose a disk", m.volumesView(), m.volumeHelp()
	case modePrompt:
		head, body, foot = "Space Explorer - scan a folder", m.promptView(), "Enter scan   Esc back"
	case modeBrowse:
		head, body, foot = "Choose a folder - "+m.br.dir, m.browseView(), m.browseHelp()
	case modeScanning:
		head, body, foot = "Scanning "+m.root, m.scanningView(), "Esc cancel"
	case modeErrors:
		head, body, foot = fmt.Sprintf("Unreadable paths (%s)", formatCount(m.summary.ErrorCount)), m.errorsView(), "arrows move   Enter reveal   Esc close"
	default:
		head, body, foot = m.treeHead(), m.treeView(), m.treeHelp()
	}
	return m.frame3(head, body, foot)
}

// frame3 lays out header, body (padded to full height), status and footer.
func (m *Model) frame3(head, body, foot string) string {
	lines := strings.Split(body, "\n")
	h := m.bodyH()
	for len(lines) < h {
		lines = append(lines, "")
	}
	status := m.status
	switch {
	case m.busy:
		status = fmt.Sprintf("Rescanning... %s files, %s (Esc cancels)", formatCount(m.progress.Files), formatSize(m.progress.Bytes))
	case m.mode == modeTree && m.status == "":
		status = m.summaryLine()
	case m.mode == modeBrowse && m.status == "":
		status = m.browseStatus()
	}
	return strings.Join([]string{
		headerStyle.Width(m.width).Render(trunc(head, m.width)),
		strings.Join(lines[:h], "\n"),
		trunc(status, m.width),
		footerStyle.Render(trunc(foot, m.width)),
	}, "\n")
}

func (m *Model) summaryLine() string {
	s := m.summary
	line := fmt.Sprintf("%s in %s files, %s folders, %s",
		formatSize(s.Size), formatCount(s.Files), formatCount(s.Dirs), formatDurationMs(s.DurationMs))
	if s.ErrorCount > 0 {
		line += errStyle.Render(fmt.Sprintf("   %s unreadable (e)", formatCount(s.ErrorCount)))
	}
	return line
}

func (m *Model) volumesView() string {
	if len(m.vols) == 0 {
		return dimStyle.Render("No disks found. Press / to type a folder path.")
	}
	var b strings.Builder
	for i, v := range m.vols {
		line := fmt.Sprintf("%-22s %s", trunc(v.Name, 22), volInfo(v.Total, v.Free, v.Path))
		if i == m.volCursor {
			line = selStyle.Width(m.width).Render("> " + line)
		} else {
			line = "  " + line
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func volInfo(total, free int64, path string) string {
	if total <= 0 {
		return path
	}
	used := total - free
	p := float64(used) * 100 / float64(total)
	return fmt.Sprintf("%s  %s used of %s", renderBar(p, 20), formatSize(used), formatSize(total))
}

func (m *Model) volumeHelp() string {
	help := "arrows choose   Enter scan   f browse   / path   q quit"
	if m.cfg.Home != "" {
		help = "arrows choose   Enter scan   f browse   h home   / path   q quit"
	}
	if len(m.rows) > 0 {
		help += "   b back to the tree"
	}
	return help
}

func (m *Model) promptView() string {
	return "Folder to scan (~ is your home):\n\n  " + m.input + "_"
}

var spinner = []string{"|", "/", "-", `\`}

func (m *Model) scanningView() string {
	p := m.progress
	return strings.Join([]string{
		"",
		fmt.Sprintf("  %s  %s files, %s folders, %s",
			spinner[m.frame%len(spinner)], formatCount(p.Files), formatCount(p.Dirs), formatSize(p.Bytes)),
		fmt.Sprintf("     %s unreadable", formatCount(p.Errors)),
		"",
		dimStyle.Render("  " + trunc(p.Current, max(10, m.width-4))),
	}, "\n")
}

func (m *Model) errorsView() string {
	var b strings.Builder
	if note := noteFor(m.errs, m.summary.ErrorCount); note != "" {
		b.WriteString(dimStyle.Render(trunc(note, m.width)) + "\n")
	}
	room := m.bodyH() - strings.Count(b.String(), "\n")
	end := min(len(m.errs), m.errTop+room)
	for i := m.errTop; i < end; i++ {
		e := m.errs[i]
		line := trunc(e.Path+"   "+e.Err, m.width-2)
		if i == m.errCursor {
			line = selStyle.Width(m.width).Render("> " + line)
		} else {
			line = "  " + errStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func trunc(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(w).Render(s)
}
