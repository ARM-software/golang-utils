package command

import (
	"fmt"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func TestCommandAsDifferentUser_Redefine(t *testing.T) {
	assert.Equal(t, "sudo test 1 2 3", Sudo().RedefineInShellForm("test", "1", "2", "3"))
	name := faker.Username()
	assert.Equal(t, fmt.Sprintf("su %v test 1 2 3", name), Su(name).RedefineInShellForm("test", "1", "2", "3"))
	name = faker.Username()
	assert.Equal(t, fmt.Sprintf("gosu %v test 1 2 3", name), Gosu(name).RedefineInShellForm("test", "1", "2", "3"))
	assert.Equal(t, "test 1 2 3", NewCommandAsDifferentUser().RedefineInShellForm("test", "1", "2", "3"))
	assert.Equal(t, "test", Me().RedefineInShellForm("test"))
	assert.Empty(t, Me().RedefineInShellForm(""))
}
