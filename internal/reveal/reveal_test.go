package reveal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShowMissingPath(t *testing.T) {
	if err := Show(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected an error for a missing path")
	}
}

// The command is built per OS but never executed here (it would pop up a
// file manager window during tests).
func TestCommandShape(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.bin")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	name, args := command(dir, true)
	if name == "" || len(args) == 0 || !strings.Contains(strings.Join(args, " "), dir) {
		t.Errorf("folder command = %q %v", name, args)
	}
	name, args = command(file, false)
	joined := strings.Join(args, " ")
	if name == "" || len(args) == 0 {
		t.Fatalf("file command = %q %v", name, args)
	}
	switch runtime.GOOS {
	case "darwin":
		if args[0] != "-R" || !strings.Contains(joined, file) {
			t.Errorf("darwin should select the file: %v", args)
		}
	case "windows":
		if !strings.HasPrefix(args[0], "/select,") {
			t.Errorf("windows should select the file: %v", args)
		}
	default:
		if args[0] != dir {
			t.Errorf("linux should open the parent folder: %v", args)
		}
	}
}
