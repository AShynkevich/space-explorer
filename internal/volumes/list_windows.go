package volumes

import (
	"golang.org/x/sys/windows"
)

// list returns the logical drives that are fixed, removable or network.
func list() []Volume {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil
	}
	var out []Volume
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		root := string(rune('A'+i)) + `:\`
		p, err := windows.UTF16PtrFromString(root)
		if err != nil {
			continue
		}
		switch windows.GetDriveType(p) {
		case windows.DRIVE_FIXED, windows.DRIVE_REMOVABLE, windows.DRIVE_REMOTE:
			out = append(out, Volume{Name: label(p, root), Path: root})
		}
	}
	return out
}

func label(root *uint16, fallback string) string {
	buf := make([]uint16, windows.MAX_PATH+1)
	if err := windows.GetVolumeInformation(root, &buf[0], uint32(len(buf)), nil, nil, nil, nil, 0); err == nil {
		if s := windows.UTF16ToString(buf); s != "" {
			return s + " (" + fallback[:2] + ")"
		}
	}
	return fallback
}

// space returns total and user-available bytes of the drive at path.
func space(path string) (total, free int64) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0
	}
	var avail, tot, totFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &avail, &tot, &totFree); err != nil {
		return 0, 0
	}
	return int64(tot), int64(avail)
}
