package command

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestCommandAsDifferentUser_Redefine(t *testing.T) {
	assert.Equal(t, "sudo test 1 2 3", Sudo().RedefineInShellForm("test", "1", "2", "3"))
	name := faker.Username()
	assert.Equal(t, fmt.Sprintf("su %v test 1 2 3", name), Su(name).RedefineInShellForm("test", "1", "2", "3"))
	name = faker.Username()
	assert.Equal(t, fmt.Sprintf("gosu %v test 1 2 3", name), Gosu(name).RedefineInShellForm("test", "1", "2", "3"))
	assert.Equal(t, fmt.Sprintf("runas /user:%v test 1 2 3", name), RunAs(name).RedefineInShellForm("test", "1", "2", "3"))
	assert.Equal(t, "elevate test", Elevate().RedefineInShellForm("test"))
	assert.Equal(t, "shellrunas /quiet test", ShellRunAs().RedefineInShellForm("test"))
	assert.Equal(t, "test 1 2 3", NewCommandAsDifferentUser().RedefineInShellForm("test", "1", "2", "3"))
	assert.Equal(t, "test", Me().RedefineInShellForm("test"))
	assert.Empty(t, Me().RedefineInShellForm(""))
	cmd, args := Me().Redefine("test")
	assert.Equal(t, "test", cmd)
	assert.Empty(t, args)
}

func TestCommandAsDifferentUser_Prepend(t *testing.T) {
	name := faker.Username()
	assert.Equal(t, fmt.Sprintf("sudo gosu %v test 1 2 3", name), Gosu(name).Prepend(Sudo()).RedefineInShellForm("test", "1", "2", "3"))
	name = faker.Username()
	assert.Equal(t, fmt.Sprintf("sudo gosu %v test", name), Gosu(name).Prepend(Sudo()).RedefineInShellForm("test"))
	cmd, args := Me().Prepend(Sudo()).Redefine("test")
	assert.Equal(t, "sudo", cmd)
	assert.Len(t, args, 1)
	cmd, args = Me().Prepend(NewCommandAsDifferentUser()).Redefine("test")
	assert.Equal(t, "test", cmd)
	assert.Empty(t, args)

}
