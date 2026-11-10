package scanner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSymlinksNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges on Windows")
	}
	root := t.TempDir()
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "huge"), 9999)
	writeFile(t, filepath.Join(root, "real"), 100)
	if err := os.Symlink(target, filepath.Join(root, "dirlink")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "huge"), filepath.Join(root, "filelink")); err != nil {
		t.Fatal(err)
	}
	// a loop: root/loop -> root
	if err := os.Symlink(root, filepath.Join(root, "loop")); err != nil {
		t.Fatal(err)
	}
	res := run(t, root, testOpts())
	if res.Root.Size != 100 {
		t.Errorf("size = %d, want 100 (symlinks must not be followed)", res.Root.Size)
	}
}

func TestHardLinksCountedOnce(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("inode identity not available")
	}
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a"), 1000)
	if err := os.Link(filepath.Join(root, "a"), filepath.Join(root, "b")); err != nil {
		t.Fatal(err)
	}
	res := run(t, root, testOpts())
	if res.Root.Size != 1000 {
		t.Errorf("size = %d, want 1000 (hard link counted twice)", res.Root.Size)
	}
}

func TestUnreadableFolder(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permissions not enforced")
	}
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "ok", "f"), 300)
	locked := filepath.Join(root, "locked")
	writeFile(t, filepath.Join(locked, "f"), 700)
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })

	res := run(t, root, testOpts())
	if res.Root.Size != 300 {
		t.Errorf("size = %d, want 300", res.Root.Size)
	}
	if res.ErrorCount != 1 || len(res.Errors) != 1 || res.Errors[0].Path != locked {
		t.Errorf("errors = %+v (count %d), want one for %s", res.Errors, res.ErrorCount, locked)
	}
	if !res.Root.Unreadable {
		t.Error("root should be flagged unreadable (propagated)")
	}
}

func TestAllocatedSizeIsBlockBased(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no block info")
	}
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "one"), 1) // occupies at least one block
	o := testOpts()
	o.Apparent = false
	res := run(t, root, o)
	if res.Root.Size < 512 {
		t.Errorf("allocated size = %d, expected at least one block", res.Root.Size)
	}
	checkInvariants(t, res.Root)
}

func TestCancel(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 50; i++ {
		writeFile(t, filepath.Join(root, "d"+string(rune('a'+i%26)), string(rune('a'+i/26)), "f"), 10)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() {
		_, err := New(root, testOpts()).Run(ctx)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("expected cancellation error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

func TestNotADirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "f")
	writeFile(t, f, 1)
	if _, err := New(f, testOpts()).Run(context.Background()); err == nil {
		t.Error("expected error for a file root")
	}
	if _, err := New(filepath.Join(f, "missing"), testOpts()).Run(context.Background()); err == nil {
		t.Error("expected error for a missing root")
	}
}
