//go:build darwin
// +build darwin

package platform

import (
	"path/filepath"
)

func determineDefaultHomeDirectory(username string) (string, error) {
	return filepath.Join(`/`, "Users", username), nil
}
