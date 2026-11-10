package platform

// SkipPaths are never descended when met at exactly these absolute paths.
// Scanning "/" on modern macOS already reaches the data volume through
// firmlinks, so /System/Volumes would count everything twice.
var SkipPaths = []string{"/System/Volumes", "/Volumes", "/dev"}
