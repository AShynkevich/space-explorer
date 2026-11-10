// Package reveal shows a file or folder in the OS file manager
// (Finder, Explorer, or the default file manager on Linux).
package reveal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Show opens the file manager at path. Files are selected in their folder
// where the OS supports it; folders are opened.
func Show(path string) error {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	name, args := command(path, info.IsDir())
	if name == "" {
		return fmt.Errorf("reveal is not supported on this platform")
	}
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }() // reap the child; explorer.exe may exit non-zero
	return nil
}
