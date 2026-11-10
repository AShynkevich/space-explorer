package reveal

// command: `explorer /select,<path>` selects the item in Explorer.
func command(path string, isDir bool) (string, []string) {
	if isDir {
		return "explorer", []string{path}
	}
	return "explorer", []string{"/select," + path}
}
