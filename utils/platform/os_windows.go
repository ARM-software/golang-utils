//go:build windows

package platform

import "golang.org/x/sys/windows/registry"

func expandFromEnvironment(s string) string {
	expanded, err := registry.ExpandString(s)
	if err == nil {
		return expanded
	}
	return s
}
