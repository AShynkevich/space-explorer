package tui

import (
	"testing"

	"space-explorer/internal/session"
)

func TestLoadShowsRootThenChildren(t *testing.T) {
	m, _ := newTree(t)
	if got, want := names(m), "/r,.a,.b,.f.bin,.3 other items"; got != want {
		t.Fatalf("rows = %q, want %q", got, want)
	}
	if m.mode != modeTree || !m.rows[0].expanded {
		t.Fatalf("mode=%v rootExpanded=%v", m.mode, m.rows[0].expanded)
	}
}

func TestExpandCollapseDropsNested(t *testing.T) {
	m, _ := newTree(t)
	press(m, "down", "right") // /r/a
	if got, want := names(m), "/r,.a,..deep,..x,.b,.f.bin,.3 other items"; got != want {
		t.Fatalf("after expand: %q", got)
	}
	press(m, "down", "right") // deep
	press(m, "up", "left")
	if got, want := names(m), "/r,.a,.b,.f.bin,.3 other items"; got != want {
		t.Fatalf("after collapse: %q", got)
	}
	press(m, "right", "enter") // re-expand then toggle off
	if m.rows[m.cursor].expanded {
		t.Fatal("enter should collapse an expanded folder")
	}
}

func TestLeftOnLeafMovesToParent(t *testing.T) {
	m, _ := newTree(t)
	press(m, "down", "right", "down", "down", "left")
	if m.rows[m.cursor].e.Name != "a" {
		t.Fatalf("cursor on %q, want parent a", m.rows[m.cursor].e.Name)
	}
}

func TestFilesAndOtherRowAreLeaves(t *testing.T) {
	m, _ := newTree(t)
	before := names(m)
	press(m, "end", "right", "enter")
	if names(m) != before {
		t.Fatalf("leaf toggled: %q", names(m))
	}
}

func TestCursorClampsAndScrolls(t *testing.T) {
	m, _ := newTree(t)
	m.height = 6 // body height 2
	press(m, "up", "up")
	if m.cursor != 0 {
		t.Fatalf("cursor = %d", m.cursor)
	}
	press(m, "end")
	if m.cursor != 4 || m.top != 3 {
		t.Fatalf("end: cursor=%d top=%d", m.cursor, m.top)
	}
	press(m, "home")
	if m.cursor != 0 || m.top != 0 {
		t.Fatalf("home: cursor=%d top=%d", m.cursor, m.top)
	}
}

func TestSortToggleRefreshesKeepingExpandedAndSelection(t *testing.T) {
	m, fb := newTree(t)
	press(m, "down", "right", "down") // a expanded, cursor on deep
	press(m, "s")
	if m.sortBy != session.SortName {
		t.Fatalf("sortBy = %q", m.sortBy)
	}
	if fb.sorts[len(fb.sorts)-1] != session.SortName {
		t.Fatalf("last Children sort = %q", fb.sorts[len(fb.sorts)-1])
	}
	if !m.rows[1].expanded || m.rows[m.cursor].e.Name != "deep" {
		t.Fatalf("lost state: %q cursor on %q", names(m), m.rows[m.cursor].e.Name)
	}
	press(m, "s")
	if m.sortBy != session.SortSize {
		t.Fatalf("sortBy = %q", m.sortBy)
	}
}

func TestOtherRowSelectionSurvivesRefresh(t *testing.T) {
	m, _ := newTree(t)
	press(m, "end", "s")
	if m.rows[m.cursor].e.Kind != session.KindOther {
		t.Fatalf("cursor on %q", m.rows[m.cursor].e.Name)
	}
}
