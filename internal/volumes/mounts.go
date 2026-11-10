package volumes

import (
	"bufio"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

var networkFS = map[string]bool{"nfs": true, "nfs4": true, "cifs": true, "smb3": true}

// parseMounts reads /proc/mounts-formatted data and keeps real block devices
// and network shares, dropping pseudo file systems and loop devices (snaps).
// It lives outside the linux-only file so it can be tested on every OS.
func parseMounts(r io.Reader) []Volume {
	var out []Volume
	seen := map[string]bool{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 3 {
			continue
		}
		dev, mount, fstype := f[0], unescape(f[1]), f[2]
		real := strings.HasPrefix(dev, "/dev/") && !strings.HasPrefix(dev, "/dev/loop")
		if !(real || networkFS[fstype]) || seen[mount] {
			continue
		}
		seen[mount] = true
		name := filepath.Base(mount)
		if mount == "/" {
			name = "Root"
		}
		out = append(out, Volume{Name: name, Path: mount})
	}
	return out
}

// unescape decodes the octal escapes (\040 = space) used in mount tables.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if v, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(v))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
