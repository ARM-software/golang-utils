package subprocess

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// INTERNAL
// wrapper over an exec cmd.
type cmdWrapper struct {
	mu  deadlock.RWMutex
	cmd *exec.Cmd
}

func (c *cmdWrapper) Set(cmd *exec.Cmd) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd == nil {
		c.cmd = cmd
	}
}

func (c *cmdWrapper) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cmd = nil
}

func (c *cmdWrapper) Start() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
	}
	return commonerrors.ConvertContextError(c.cmd.Start())
}

func (c *cmdWrapper) Run() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
	}
	return commonerrors.ConvertContextError(c.cmd.Run())
}

func (c *cmdWrapper) Stop() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
	}
	subprocess := c.cmd.Process
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if subprocess != nil {
		pid := subprocess.Pid
		parallelisation.ScheduleAfter(ctx, 10*time.Millisecond, func(time.Time) {
			process, err := os.FindProcess(pid)
			if process == nil || err != nil {
				return
			}
			_ = process.Kill()
		})
	}
	_ = c.cmd.Wait()
	return nil
}

func (c *cmdWrapper) Pid() (pid int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		err = fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
		return
	}
	subprocess := c.cmd.Process
	if subprocess == nil {
		err = fmt.Errorf("%w:undefined subprocess", commonerrors.ErrUndefined)
		return
	}
	pid = subprocess.Pid
	return
}

// Definition of a command
type command struct {
	cmd        string
	args       []string
	loggers    logs.Loggers
	cmdWrapper cmdWrapper
}

func (c *command) createCommand(cmdCtx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(cmdCtx, c.cmd, c.args...) //nolint:gosec
	cmd.Stdout = newOutStreamer(c.loggers)
	cmd.Stderr = newErrLogStreamer(c.loggers)
	return cmd
}

func (c *command) GetPath() string {
	return c.cmd
}

func (c *command) GetCmd(cmdCtx context.Context) *cmdWrapper {
	c.cmdWrapper.Set(c.createCommand(cmdCtx))
	return &c.cmdWrapper
}

func (c *command) Reset() {
	c.cmdWrapper.Reset()
}

func (c *command) Check() (err error) {
	if c.cmd == "" {
		err = fmt.Errorf("missing command: %w", commonerrors.ErrUndefined)
		return
	}
	if c.loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return
}

func newCommand(loggers logs.Loggers, cmd string, args ...string) (osCmd *command) {
	osCmd = &command{
		cmd:        cmd,
		args:       args,
		loggers:    loggers,
		cmdWrapper: cmdWrapper{},
	}
	return
}
