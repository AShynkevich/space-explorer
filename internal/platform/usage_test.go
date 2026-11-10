package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUsageOf(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(p)
	if err != nil {
		t.Fatal(err)
	}

	if u := UsageOf(info, true); u.Size != 1000 {
		t.Errorf("apparent size = %d, want 1000", u.Size)
	}
	u := UsageOf(info, false)
	if runtime.GOOS == "windows" {
		if u.Size != 1000 || u.HasID {
			t.Errorf("windows fallback = %+v", u)
		}
		return
	}
	if !u.HasID || u.Nlink != 1 || u.Ino == 0 {
		t.Errorf("identity not filled: %+v", u)
	}
	if u.Size < 1000 || u.Size%512 != 0 {
		t.Errorf("allocated size = %d, want a multiple of 512 and >= 1000", u.Size)
	}
}
