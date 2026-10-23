package util

import "strings"

// IsPathComponent reports whether a name is one portable file component.
func IsPathComponent(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\:\x00")
}
