package subprocess

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARMmbed/golang-utils/utils/logs"
	"github.com/ARMmbed/golang-utils/utils/platform"
)

func TestExecute(t *testing.T) {
	var loggers logs.Loggers = &logs.GenericLoggers{}
	err := loggers.Check()
	assert.NotNil(t, err)

	err = Execute(loggers, "", "", "", "ls")
	assert.NotNil(t, err)

	loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	if platform.IsWindows() {
		err = Execute(loggers, "", "", "", "cmd", "dir")
	} else {
		err = Execute(loggers, "", "", "", "ls", "-l")
	}
	require.Nil(t, err)
}

func TestShortLivedSubprocess(t *testing.T) {
	var loggers logs.Loggers = &logs.GenericLoggers{}
	err := loggers.Check()
	assert.NotNil(t, err)

	_, err = Create(loggers, "", "", "", "ls")
	assert.NotNil(t, err)

	loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	var p *Subprocess
	if platform.IsWindows() {
		p, err = Create(loggers, "", "", "", "cmd", "dir")
	} else {
		p, err = Create(loggers, "", "", "", "ls", "-l")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()
	_testSubprocess(t, p)
}

func TestLongerLivedSubprocess(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)

	var p *Subprocess
	if platform.IsWindows() {
		p, err = Create(loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = Create(loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()
	_testSubprocess(t, p)
}

func _testSubprocess(t *testing.T, p *Subprocess) {
	assert.False(t, p.IsOn())
	err := p.Start()
	require.Nil(t, err)
	assert.True(t, p.IsOn())

	//Checking idempotence
	err = p.Start()
	require.Nil(t, err)
	err = p.Check()
	require.Nil(t, err)

	time.Sleep(time.Duration(200) * time.Millisecond)

	err = p.Restart()
	require.Nil(t, err)
	assert.True(t, p.IsOn())

	err = p.Stop()
	require.Nil(t, err)
	assert.False(t, p.IsOn())

	//Checking idempotence
	err = p.Stop()
	require.Nil(t, err)
	err = p.Execute()
	assert.NotNil(t, err)
}
