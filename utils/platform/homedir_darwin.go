//go:build darwin

package platform

import "fmt"

func determineDefaultHomeDirectory(username string) (string, error) {
	return fmt.Sprintf("/Users/%v", username), nil
}
