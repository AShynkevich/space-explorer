package main

import (
	"context"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"space-explorer/internal/reveal"
	"space-explorer/internal/scanner"
	"space-explorer/internal/session"
	"space-explorer/internal/volumes"
)

// App is the object bound to the frontend. Its exported methods become
// callable from JavaScript (frontend/wailsjs/go/main/App.js). It is glue
// only: all logic lives in internal/*.
type App struct {
	ctx      context.Context
	sessions *session.Manager
}

// NewApp creates the App.
func NewApp() *App {
	return &App{}
}

// startup keeps the Wails context and wires scan events to the frontend.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.sessions = session.New(func(event string, data any) {
		runtime.EventsEmit(ctx, event, data)
	}, scanner.DefaultOptions())
}

// shutdown stops a running scan when the window closes.
func (a *App) shutdown(context.Context) {
	a.sessions.Cancel()
}

// Volumes lists the disks the user can scan.
func (a *App) Volumes() []volumes.Volume {
	return volumes.List()
}

// HomeDir returns the user's home folder (a handy default scan root).
func (a *App) HomeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// PickFolder opens the native folder dialog; returns "" if cancelled.
func (a *App) PickFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose a folder or disk to analyse",
	})
}

// StartScan begins scanning root; progress and completion arrive as events
// (see session.Event* constants).
func (a *App) StartScan(root string) error {
	return a.sessions.Start(root)
}

// CancelScan stops the running scan.
func (a *App) CancelScan() {
	a.sessions.Cancel()
}

// IsScanning reports whether a scan is in progress.
func (a *App) IsScanning() bool {
	return a.sessions.Running()
}

// Children returns the entries of a scanned folder (lazy tree).
// sortBy is "size" (default) or "name".
func (a *App) Children(path, sortBy string) (*session.Listing, error) {
	return a.sessions.Children(path, sortBy)
}

// Rescan re-reads one scanned folder and updates the tree in place.
func (a *App) Rescan(path string) error {
	return a.sessions.Rescan(path)
}

// Reveal shows path in Finder / Explorer / the Linux file manager.
func (a *App) Reveal(path string) error {
	return reveal.Show(path)
}

// ScanErrors lists the paths that could not be read in the last scan.
func (a *App) ScanErrors() []session.ErrorInfo {
	return a.sessions.Errors()
}
