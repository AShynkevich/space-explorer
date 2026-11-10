//go:build !darwin && !linux && !windows

package volumes

func space(string) (total, free int64) { return 0, 0 }

func list() []Volume { return []Volume{{Name: "Root", Path: "/"}} }
