// Command space-explorer-tui is the terminal version of Space Explorer. It
// uses the same scanner and session as the desktop app.
//
//	space-explorer-tui [folder]
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"space-explorer/internal/reveal"
	"space-explorer/internal/scanner"
	"space-explorer/internal/session"
	"space-explorer/internal/tui"
	"space-explorer/internal/volumes"
)

func main() {
	if len(os.Args) > 2 || (len(os.Args) == 2 && os.Args[1][0] == '-') {
		fmt.Fprintln(os.Stderr, "usage: space-explorer-tui [folder]")
		os.Exit(2)
	}
	cfg := tui.Config{Bridge: tui.NewBridge(), Volumes: volumes.List, Reveal: reveal.Show}
	cfg.Home, _ = os.UserHomeDir()
	if len(os.Args) == 2 {
		cfg.Start = os.Args[1]
	}
	cfg.Backend = session.New(cfg.Bridge.Emit, scanner.DefaultOptions())

	p := tea.NewProgram(tui.New(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
