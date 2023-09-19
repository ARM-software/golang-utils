package platform

import (
	"context"
	"os/user"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestUser() (testUser *user.User) {
	name := faker.Name()
	testUser = &user.User{
		Uid:      "123",
		Gid:      strings.ToLower(faker.Word()),
		Username: strings.ToLower(strings.ReplaceAll(name, " ", "_")),
		Name:     strings.ReplaceAll(name, " ", "_"),
		HomeDir:  "",
	}
	return
}
func TestDefineUser(t *testing.T) {
	// Note: on Windows, it is necessary to run this test with elevated privileges https://github.com/iamacarpet/go-win64api/issues/26
	if IsWindows() {
		t.Log("Note: it is necessary to run this test with elevated privileges https://github.com/iamacarpet/go-win64api/issues/26")
	}
	user := generateTestUser()
	err := DefineUser(context.TODO(), user, "")
	require.NoError(t, err)
	found, err := HasUser(user.Username)
	assert.NoError(t, err)
	assert.True(t, found)
	defer func() { _ = DeleteUser(context.Background(), user) }()
	err = AddGroup(context.TODO(), user.Gid)
	require.NoError(t, err)
	found, err = HasGroup(user.Gid)
	assert.NoError(t, err)
	assert.True(t, found)
	defer func() { _ = RemoveGroup(context.Background(), user.Gid) }()
	require.NoError(t, AssociateUserToGroup(context.TODO(), user.Username, user.Gid))
	require.NoError(t, DissociateUserFromGroup(context.TODO(), user.Username, user.Gid))
	require.NoError(t, DeleteUser(context.TODO(), user))
	found, err = HasUser(user.Username)
	assert.NoError(t, err)
	assert.False(t, found)
	require.NoError(t, RemoveGroup(context.TODO(), user.Gid))
	found, err = HasGroup(user.Gid)
	assert.NoError(t, err)
	assert.False(t, found)
}
