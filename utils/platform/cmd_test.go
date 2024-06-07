package platform

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

func TestWithPrivileges(t *testing.T) {
	admin, err := IsCurrentUserAnAdmin()
	require.NoError(t, err)
	name := faker.Username()
	cmd := WithPrivileges(command.Gosu(name)).RedefineInShellForm("test", "1", "2", "3")
	if admin {
		assert.Equal(t, fmt.Sprintf("gosu %v test 1 2 3", name), cmd)
	} else {
		assert.True(t, strings.Contains(cmd, fmt.Sprintf("gosu %v test 1 2 3", name)))
		assert.False(t, strings.HasPrefix(cmd, fmt.Sprintf("gosu %v test 1 2 3", name)))
	}
	cmd = WithPrivileges(nil).RedefineInShellForm("test", "1", "2", "3")
	if admin {
		assert.Equal(t, "test 1 2 3", cmd)
	} else {
		assert.True(t, strings.Contains(cmd, "test 1 2 3"))
		assert.False(t, strings.HasPrefix(cmd, "test 1 2 3"))
	}
}
