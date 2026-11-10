package reveal

// command: `open -R` selects the item in Finder, `open` opens a folder.
func command(path string, isDir bool) (string, []string) {
	if isDir {
		return "open", []string{path}
	}
	return "open", []string{"-R", path}
}
