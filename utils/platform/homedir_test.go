package platform

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultHomeDirectory(t *testing.T) {
	username := faker.Username()
	home, err := GetDefaultHomeDirectory(username)
	require.NoError(t, err)
	assert.NotEmpty(t, home)
	assert.Contains(t, home, username)
}

func TestGetHomeDirectory(t *testing.T) {
	username := faker.Username()
	home, err := GetHomeDirectory(username)
	require.NoError(t, err)
	assert.NotEmpty(t, home)
	assert.Contains(t, home, username)

	currentUser, err := GetCurrentUser()
	require.NoError(t, err)
	home, err = GetHomeDirectory(currentUser.Username)
	fmt.Println(home)
	require.NoError(t, err)
	assert.NotEmpty(t, home)

}
