// Package tui is the terminal front-end: a Bubble Tea model over the same
// session.Manager the Wails GUI uses. It holds no scan logic.
package tui

import (
	"space-explorer/internal/session"
	"space-explorer/internal/volumes"
)

// Backend is the part of session.Manager the UI needs (faked in tests).
type Backend interface {
	Start(root string) error
	Cancel()
	Rescan(path string) error
	Children(path, sortBy string) (*session.Listing, error)
	Errors() []session.ErrorInfo
}

// Config wires the model to its environment.
type Config struct {
	Backend Backend
	Bridge  *Bridge
	Volumes func() []volumes.Volume
	Reveal  func(path string) error
	Home    string // used for "h" and "~" in the path prompt
	Start   string // if set, scan this path right away
}
