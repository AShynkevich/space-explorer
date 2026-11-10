//go:build !darwin && !linux

package platform

// SkipPaths lists absolute paths that are never descended.
var SkipPaths []string
