package idgen

import "github.com/gofrs/uuid"

// Generates a UUID.
func GenerateUuid4() (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}
