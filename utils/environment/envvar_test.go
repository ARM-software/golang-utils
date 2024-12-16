package environment

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestEnvVar_Validate(t *testing.T) {
	tests := []struct {
		key         string
		value       string
		name        string
		valueRules  []validation.Rule
		expectError bool
	}{
		{
			name:        "variable with empty key",
			expectError: true,
		},
		{
			key:         faker.Sentence(),
			name:        "variable with whitespaces",
			expectError: true,
		},
		{
			key:         faker.Name(),
			name:        "variable with whitespaces",
			expectError: true,
		},

		{
			key:         faker.Word() + "=" + faker.Word(),
			name:        "variable with `=``",
			expectError: true,
		},
		{
			key:         faker.Word() + "$",
			name:        "variable with special character",
			expectError: true,
		},
		{
			key:         "0" + faker.Word(),
			name:        "variable starting with a digit",
			expectError: true,
		},
		{
			key:         faker.Word(),
			name:        "valid variable",
			expectError: false,
		},
		{
			key:         faker.Word() + "_0",
			name:        "valid variable with digit & underscore",
			expectError: false,
		},
		{
			key:         "_" + faker.Word(),
			name:        "variable starting with an underscore",
			expectError: false,
		},
		{
			key:         faker.Word(),
			name:        "variable value compliant with one rule",
			value:       faker.Sentence(),
			valueRules:  []validation.Rule{validation.Required},
			expectError: false,
		},
		{
			key:         faker.Word(),
			name:        "variable value compliant with several rules",
			value:       faker.Word(),
			valueRules:  []validation.Rule{validation.Required, is.Alphanumeric},
			expectError: false,
		},
		{
			key:         faker.Word(),
			name:        "non compliant variable value",
			valueRules:  []validation.Rule{validation.Required},
			expectError: true,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var env IEnvironmentVariable
			if len(test.valueRules) == 0 {
				env = NewEnvironmentVariable(test.key, test.value)
			} else {
				env = NewEnvironmentVariableWithValidation(test.key, test.value, test.valueRules...)
			}
			if test.expectError {
				require.Error(t, env.Validate())
			} else {
				require.NoError(t, env.Validate())
			}
			assert.Equal(t, test.key, env.GetKey())
			assert.Equal(t, test.value, env.GetValue())
		})
	}
	require.Error(t, IsEnvironmentVariableKey.Validate(faker.Sentence()))

}

func TestValidateEnvironmentVariables(t *testing.T) {
	require.NoError(t, ValidateEnvironmentVariables())
	require.NoError(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Word(), "")))
	require.NoError(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Word(), ""), NewEnvironmentVariable(faker.Word(), "")))
	require.Error(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Name(), ""), NewEnvironmentVariable(faker.Word(), "")))
	require.Error(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Word(), ""), NewEnvironmentVariable(faker.Name(), "")))
}

func TestParseEnvironmentVariable(t *testing.T) {
	env, err := ParseEnvironmentVariable(faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Nil(t, env)
	_, err = ParseEnvironmentVariable(faker.Word() + "$=" + faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	_, err = ParseEnvironmentVariable(faker.Word() + "=" + faker.Word())
	require.NoError(t, err)
	_, err = ParseEnvironmentVariable(faker.Word() + "=" + faker.Word() + "=")
	require.NoError(t, err)
	_, err = ParseEnvironmentVariable(faker.Word() + "=" + faker.Word() + "=" + faker.Sentence())
	require.NoError(t, err)
	key := strings.ReplaceAll(strings.ReplaceAll(faker.Sentence(), " ", "_"), ".", "")
	value := faker.Sentence()
	envTest := NewEnvironmentVariable(key, value)
	envTest2 := envTest
	assert.True(t, envTest.Equal(envTest2))
	assert.True(t, envTest2.Equal(envTest))
	env, err = ParseEnvironmentVariable(envTest.String())
	require.NoError(t, err)
	assert.Equal(t, key, env.GetKey())
	assert.Equal(t, value, env.GetValue())
	assert.True(t, env.Equal(envTest))
	txt, err := envTest.MarshalText()
	require.NoError(t, err)
	require.NoError(t, env.UnmarshalText(txt))
	assert.True(t, envTest.Equal(env))
}

func TestEnvVar_ParseEnvironmentVariables(t *testing.T) {
	username := faker.Username()
	entries := []string{"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/65357/bus", "HOME=/home/josjen01", faker.UUIDHyphenated(), "EDITOR=hx", "LOGNAME=josjen01", "DISPLAY=:0", "SSH_AUTH_SOCK=/tmp/ssh-eBrdhiWnaFYp/agent.4969", "KRB5CCNAME=FILE:/tmp/krb5cc_65357_XLwjEE", "GPG_AGENT_INFO=/run/user/65357/gnupg/S.gpg-agent:0:1", "LANGUAGE=en_US:", "USER=" + username, "XDG_RUNTIME_DIR=/run/user/65357", "WINDOWID=54525966", "KITTY_PID=151539", "CMSIS_PACK_ROOT=/home/josjen01/.cache/arm/packs", "XDG_SESSION_ID=4", "XDG_CONFIG_DIRS=/etc/xdg/xdg-i3:/etc/xdg", faker.Name(), "GDMSESSION=i3", "WINDOWPATH=2", "SHLVL=1", "DESKTOP_SESSION=i3", "GTK_MODULES=gail:atk-bridge", "LANG=en_US.UTF-8", "FZF_DEFAULT_OPTS=--colour dark,hl:#d65d08,hl+:#d65d08,fg+:#282828,bg+:#282828,fg+:#b58900,info:#ebdbb2,prompt:#268bd2,pointer:#2aa198,marker:#d33682,spinner:#268bd2 -m", "XDG_SESSION_DESKTOP=i3", "XDG_SESSION_TYPE=x11", "KITTY_PUBLIC_KEY=1:^1R-7)Aw|}io+D^KqaYVJF0R&a!f&dpX}gSSEIH&", "XDG_SEAT=seat0", "TERM=xterm-kitty", "XDG_DATA_DIRS=/usr/share/i3:/usr/local/share/:/usr/share/:/var/lib/snapd/desktop", "DESKTOP_STARTUP_ID=i3/kitty/4969-5-e126332_TIME31895328", "SHELL=/bin/bash", "KITTY_WINDOW_ID=6", "QT_ACCESSIBILITY=1", "COLOURTERM=truecolour", "TERMINFO=/home/josjen01/.local/kitty.app/lib/kitty/terminfo", "SSH_AGENT_PID=5033", "XDG_SESSION_CLASS=user", "KITTY_INSTALLATION_DIR=/home/josjen01/.local/kitty.app/lib/kitty", "XDG_CURRENT_DESKTOP=i3", "MANPATH=:/home/josjen01/.local/kitty.app/share/man::/opt/puppetlabs/puppet/share/man", "XAUTHORITY=/run/user/65357/gdm/Xauthority", "CMSIS_COMPILER_ROOT=/etc/cmsis-build", "XDG_VTNR=2", "I3SOCK=/run/user/65357/i3/ipc-socket.4969", "USERNAME=josjen01", "PATH=/usr/local/go/bin:/bin:/home/josjen01/.local/bin:/home/josjen01/go/bin:/home/josjen01/.cargo/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/usr/games:/usr/local/games:/snap/bin:/opt/puppetlabs/bin", "PWD=/home/josjen01/Git/golang-utils/utils/environment", "GOCOVERDIR=/tmp/go-build1026478377/b001/gocoverdir", "test1=Accusantium voluptatem aut sit perferendis consequatur", "test2=Perferendis aut accusantium voluptatem sit consequatur.", faker.Word()}
	environmentVariables := ParseEnvironmentVariables(entries...)
	assert.NotEmpty(t, environmentVariables)
	assert.Len(t, environmentVariables, len(entries)-3)
	env, err := FindEnvironmentVariable("USER", environmentVariables...)
	require.NoError(t, err)
	assert.Equal(t, username, env.GetValue())
	_, err = FindEnvironmentVariable("TEST1", environmentVariables...)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}

func TestFindEnvironmentVariable(t *testing.T) {
	entries := []string{"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/65357/bus", "HOME=first", "home=second", faker.UUIDHyphenated(), "EDITOR=hx", "logName=josjen01", "LOGNAME=josjen01", "teSt1=Accusantium voluptatem aut sit perferendis consequatur", "TEST1=Perferendis aut accusantium voluptatem sit consequatur.", faker.Word()}
	environmentVariables := ParseEnvironmentVariables(entries...)
	home, err := FindEnvironmentVariable("HOME", environmentVariables...)
	require.NoError(t, err)
	assert.Equal(t, "first", home.GetValue())
	home, err = FindEnvironmentVariable("home", environmentVariables...)
	require.NoError(t, err)
	assert.Equal(t, "second", home.GetValue())
	home, err = FindEnvironmentVariable(faker.Username(), environmentVariables...)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.Empty(t, home)
	home, err = FindFoldEnvironmentVariable(faker.Username(), environmentVariables...)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.Empty(t, home)
	home, err = FindFoldEnvironmentVariable("home", environmentVariables...)
	require.NoError(t, err)
	assert.Equal(t, "first", home.GetValue())
	test1, err := FindEnvironmentVariable("TEST1", environmentVariables...)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(test1.GetValue(), "Perferendis"))
	test1, err = FindFoldEnvironmentVariable("TEST1", environmentVariables...)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(test1.GetValue(), "Accusantium"))
}
func TestExpandEnvironmentVariable(t *testing.T) {
	env1 := NewEnvironmentVariable("test", platform.SubstituteParameter("test"))
	expanded1 := ExpandEnvironmentVariable(true, env1)
	require.NotEmpty(t, expanded1)
	assert.True(t, env1.Equal(expanded1))
	expanded1 = ExpandEnvironmentVariable(true, env1, env1)
	require.NotEmpty(t, expanded1)
	assert.True(t, env1.Equal(expanded1))
	env2 := NewEnvironmentVariable("test2", platform.SubstituteParameter("test3"))
	env3 := NewEnvironmentVariable("test3", platform.SubstituteParameter("test"))
	expanded2 := ExpandEnvironmentVariable(true, env2, env1, env2, env3)
	require.NotEmpty(t, expanded2)
	assert.False(t, env1.Equal(expanded2))
	assert.Equal(t, expanded2.GetValue(), env1.GetValue())
}

func TestExpandEnvironmentVariables(t *testing.T) {
	username := faker.Username()
	entries := []string{fmt.Sprintf("DBUS_SESSION_BUS_ADDRESS=system:path=/run%v/65357/bus/", platform.SubstituteParameter("HOME")), fmt.Sprintf("HOME=/home/%v", platform.SubstituteParameter("LOGNAME")), fmt.Sprintf("LOGNAME=%v", username)}
	expandedEnvironmentVariables := ExpandEnvironmentVariables(true, ParseEnvironmentVariables(entries...)...)
	require.NotEmpty(t, expandedEnvironmentVariables)
	logname, err := FindEnvironmentVariable("LOGNAME", expandedEnvironmentVariables...)
	require.NoError(t, err)
	assert.Equal(t, username, logname.GetValue())
	dbus, err := FindEnvironmentVariable("DBUS_SESSION_BUS_ADDRESS", expandedEnvironmentVariables...)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("system:path=/run/home/%v/65357/bus/", username), dbus.GetValue())
}

func TestSortEnvironmentVariables(t *testing.T) {
	entries := []string{fmt.Sprintf("ccc%v=%v", faker.Username(), faker.Sentence()), fmt.Sprintf("Aaodasdoah%v=%v", faker.Username(), faker.Sentence()), "b=second", fmt.Sprintf("Za%v=%v", faker.Word(), faker.Sentence())}
	envVars := ParseEnvironmentVariables(entries...)
	SortEnvironmentVariables(envVars)
	require.Len(t, envVars, 4)
	assert.True(t, strings.HasPrefix(envVars[0].String(), "A"))
	assert.True(t, strings.HasPrefix(envVars[1].String(), "b"))
	assert.True(t, strings.HasPrefix(envVars[2].String(), "c"))
	assert.True(t, strings.HasPrefix(envVars[3].String(), "Z"))

	var empty []IEnvironmentVariable
	SortEnvironmentVariables(empty)
}

func TestUniqueEnvironmentVariables(t *testing.T) {
	randomKey := faker.Word()
	entries := []string{"DBUS_SESSION_BUS_ADDRESS=system:path=/run/65357/bus/", "HOME=first", "home=second", fmt.Sprintf("%v=%v", randomKey, faker.Sentence())}
	envVars := ParseEnvironmentVariables(entries...)
	require.NotEmpty(t, envVars)
	assert.Len(t, envVars, 4)
	assert.Empty(t, UniqueEnvironmentVariables(true))
	uniqueEnvVars := UniqueEnvironmentVariables(false, envVars...)
	require.NotEmpty(t, uniqueEnvVars)
	assert.NotEqual(t, uniqueEnvVars, envVars)
	assert.Len(t, uniqueEnvVars, 3)
	home, err := FindEnvironmentVariable("HOME", uniqueEnvVars...)
	require.NoError(t, err)
	assert.Equal(t, "first", home.GetValue())

	uniqueEnvVars2 := UniqueEnvironmentVariables(false, uniqueEnvVars...)
	require.NotEmpty(t, uniqueEnvVars2)
	SortEnvironmentVariables(uniqueEnvVars2)
	SortEnvironmentVariables(uniqueEnvVars)
	assert.EqualValues(t, uniqueEnvVars2, uniqueEnvVars)

	uniqueEnvVars3 := UniqueEnvironmentVariables(true, envVars...)
	require.NotEmpty(t, uniqueEnvVars3)
	SortEnvironmentVariables(uniqueEnvVars3)
	assert.Len(t, uniqueEnvVars3, 4)
	assert.NotEqualValues(t, uniqueEnvVars3, uniqueEnvVars)

}

func TestMergeEnvironmentVariables(t *testing.T) {
	randomKey := fmt.Sprintf("B%v", faker.Word())
	entries1 := []string{fmt.Sprintf("ccc%v=%v", faker.Username(), faker.Sentence()), "HOME=first", "home=second", fmt.Sprintf("%v=%v", randomKey, faker.Sentence())}
	entries2 := []string{"Zabcd=tmp", "HOME=third", fmt.Sprintf("%v=%v", randomKey, faker.Sentence())}
	envVars := MergeEnvironmentVariableSets(false, ParseEnvironmentVariables(entries1...), ParseEnvironmentVariables(entries2...)...)
	require.NotEmpty(t, envVars)
	assert.Len(t, envVars, 4)
	home, err := FindEnvironmentVariable("HOME", envVars...)
	require.NoError(t, err)
	assert.Equal(t, "first", home.GetValue())
	home, err = FindEnvironmentVariable("home", envVars...)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.Empty(t, home)

	uniqueEnvVars := UniqueEnvironmentVariables(false, envVars...)
	require.NotEmpty(t, uniqueEnvVars)
	SortEnvironmentVariables(envVars)
	SortEnvironmentVariables(uniqueEnvVars)
	assert.EqualValues(t, envVars, uniqueEnvVars)
	assert.True(t, strings.HasPrefix(envVars[0].String(), "B"))
	assert.True(t, strings.HasPrefix(envVars[1].String(), "c"))
	assert.True(t, strings.HasPrefix(envVars[2].String(), "H"))
	assert.True(t, strings.HasPrefix(envVars[3].String(), "Z"))

	envVars = MergeEnvironmentVariableSets(true, ParseEnvironmentVariables(entries1...), ParseEnvironmentVariables(entries2...)...)
	require.NotEmpty(t, envVars)
	assert.Len(t, envVars, 5)
	home, err = FindEnvironmentVariable("HOME", envVars...)
	require.NoError(t, err)
	assert.Equal(t, "first", home.GetValue())
	home, err = FindEnvironmentVariable("home", envVars...)
	require.NoError(t, err)
	assert.Equal(t, "second", home.GetValue())
}

func TestCloneEnvironmentVariable(t *testing.T) {
	env1 := NewEnvironmentVariable(faker.Word(), faker.Sentence())
	clone1 := CloneEnvironmentVariable(env1)
	assert.Equal(t, env1, env1)
	assert.False(t, clone1 == env1)
	assert.True(t, clone1.Equal(env1))
	assert.True(t, env1.Equal(clone1))
	clone2 := CloneEnvironmentVariable(nil)
	assert.Nil(t, clone2)
}
