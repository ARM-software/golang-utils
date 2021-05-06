package subprocess

import (
	"context"
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

	err = Execute(context.Background(), loggers, "", "", "", "ls")
	assert.NotNil(t, err)

	loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	if platform.IsWindows() {
		err = Execute(context.Background(), loggers, "", "", "", "cmd", "/c", "dir")
	} else {
		err = Execute(context.Background(), loggers, "", "", "", "ls", "-l")
	}
	require.Nil(t, err)
}

func TestShortLivedSubprocess(t *testing.T) {
	var loggers logs.Loggers = &logs.GenericLoggers{}
	err := loggers.Check()
	assert.NotNil(t, err)

	_, err = New(context.Background(), loggers, "", "", "", "ls")
	assert.NotNil(t, err)

	loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(context.Background(), loggers, "", "", "", "cmd", "dir")
	} else {
		p, err = New(context.Background(), loggers, "", "", "", "ls", "-l")
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
		p, err = New(context.Background(), loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(context.Background(), loggers, "", "", "", "sleep", "4")
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

func TestCancelledLongerLivedSubprocess(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	cancellableCtx, cancelFunc := context.WithCancel(context.Background())

	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(cancellableCtx, loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(cancellableCtx, loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()

	assert.False(t, p.IsOn())
	err = p.Start()
	require.Nil(t, err)
	assert.True(t, p.IsOn())
	time.Sleep(10 * time.Millisecond)
	cancelFunc()
	time.Sleep(200 * time.Millisecond)
	assert.False(t, p.IsOn())
}

func TestCancelledLongerLivedSubprocess2(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	ctx := context.Background()
	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(ctx, loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(ctx, loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()

	ready := make(chan bool)
	go func() {
		ready <- true
		_ = p.Execute()
	}()
	<-ready
	time.Sleep(10 * time.Millisecond)
	assert.True(t, p.IsOn())
	time.Sleep(10 * time.Millisecond)
	p.Cancel()
	time.Sleep(200 * time.Millisecond)
	assert.False(t, p.IsOn())
}

func TestCancelledLongerLivedSubprocess3(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	ctx, cancelFunc := context.WithCancel(context.Background())
	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(ctx, loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(ctx, loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()

	ready := make(chan bool)
	go func() {
		ready <- true
		_ = p.Execute()
	}()
	<-ready
	time.Sleep(10 * time.Millisecond)
	assert.True(t, p.IsOn())
	time.Sleep(10 * time.Millisecond)
	cancelFunc()
	time.Sleep(200 * time.Millisecond)
	assert.False(t, p.IsOn())
}
