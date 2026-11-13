package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
	footerStyle = lipgloss.NewStyle().Faint(true)
	dimStyle    = lipgloss.NewStyle().Faint(true)
	dirStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	otherStyle  = lipgloss.NewStyle().Faint(true).Italic(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	selStyle    = lipgloss.NewStyle().Background(lipgloss.Color("238")).Bold(true)
)

// barColor goes green (small) -> yellow -> red (>= 50% of the root),
// the same ramp as the GUI: hsl(130 -> 0, 70%, 48%).
func barColor(percent float64) lipgloss.Color {
	p := math.Min(math.Max(percent, 0), 100)
	hue := math.Round(130 * (1 - math.Min(p/50, 1)))
	return lipgloss.Color(hslHex(hue, 0.70, 0.48))
}

func hslHex(h, s, l float64) string {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	default:
		r, g, b = 0, c, x
	}
	return fmt.Sprintf("#%02x%02x%02x", int(math.Round((r+m)*255)), int(math.Round((g+m)*255)), int(math.Round((b+m)*255)))
}

// renderBar draws a w-cell bar filled to p percent (at least one cell when p > 0).
func renderBar(p float64, w int) string {
	if w <= 0 {
		return ""
	}
	n := int(math.Round(p / 100 * float64(w)))
	if p > 0 && n == 0 {
		n = 1
	}
	n = min(n, w)
	on := lipgloss.NewStyle().Foreground(barColor(p)).Render(strings.Repeat("█", n))
	return on + dimStyle.Render(strings.Repeat("░", w-n))
}

func displayWidth(s string) int { return lipgloss.Width(s) }
