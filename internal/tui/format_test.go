package tui

import (
	"testing"
	"time"
)

func TestFormatSize(t *testing.T) {
	for in, want := range map[int64]string{0: "0 B", -5: "0 B", 1023: "1023 B", 1024: "1.0 KB",
		1536: "1.5 KB", 1 << 20: "1.0 MB", 5 << 30: "5.0 GB", 1048575: "1.0 MB"} {
		if got := formatSize(in); got != want {
			t.Errorf("formatSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatPercentCountDuration(t *testing.T) {
	if formatPercent(0) != "0%" || formatPercent(0.05) != "<0.1%" || formatPercent(12.34) != "12.3%" || formatPercent(99.99) != "100%" {
		t.Error("formatPercent")
	}
	if formatCount(1234567) != "1,234,567" || formatCount(12) != "12" || formatCount(1000) != "1,000" {
		t.Error("formatCount")
	}
	if formatDuration(500*time.Millisecond) != "500 ms" || formatDuration(2500*time.Millisecond) != "2.5 s" ||
		formatDuration(125*time.Second) != "2 min 5 s" || formatDuration(119600*time.Millisecond) != "2 min 0 s" {
		t.Error("formatDuration")
	}
}

func TestBarColorRamp(t *testing.T) {
	if barColor(0) != "#26d946" && barColor(0) == barColor(50) {
		t.Error("ramp")
	}
	if barColor(60) != barColor(50) || barColor(-3) != barColor(0) {
		t.Error("clamp")
	}
}

func TestRenderBar(t *testing.T) {
	if displayWidth(renderBar(0, 10)) != 10 || displayWidth(renderBar(100, 10)) != 10 || renderBar(5, 0) != "" {
		t.Error("width")
	}
}
