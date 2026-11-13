package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var units = []string{"B", "KB", "MB", "GB", "TB", "PB"}

// formatSize: 1536 -> "1.5 KB" (binary units, like the scanner's du-style sizes).
func formatSize(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	v, i := float64(b), 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	if fmt.Sprintf("%.1f", v) == "1024.0" && i < len(units)-1 {
		v, i = 1, i+1
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

func formatPercent(p float64) string {
	if !(p > 0) {
		return "0%"
	}
	if p < 0.1 {
		return "<0.1%"
	}
	s := fmt.Sprintf("%.1f", p)
	if s == "100.0" {
		return "100%"
	}
	return s + "%"
}

// formatCount: 1234567 -> "1,234,567".
func formatCount(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 && c != '-' {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	if ms < 60000 {
		return fmt.Sprintf("%.1f s", float64(ms)/1000)
	}
	m, s := ms/60000, int(math.Round(float64(ms%60000)/1000))
	if s == 60 {
		m, s = m+1, 0
	}
	return fmt.Sprintf("%d min %d s", m, s)
}
