package volumes

import "os"

// list reads the mount table; see parseMounts for the filtering rules.
func list() []Volume {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return []Volume{{Name: "Root", Path: "/"}}
	}
	defer f.Close()
	vs := parseMounts(f)
	if len(vs) == 0 {
		return []Volume{{Name: "Root", Path: "/"}}
	}
	return vs
}
