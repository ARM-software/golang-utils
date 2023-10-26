//go:build windows
// +build windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func determineDefaultHomeDirectory(username string) (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	if drive == "" {
		return "", fmt.Errorf("%w: cannot determine the default home drive", commonerrors.ErrUnexpected)
	}
	return filepath.Join(drive, `\`, "Users", username), nil
}
