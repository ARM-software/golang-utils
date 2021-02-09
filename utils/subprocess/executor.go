package subprocess

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
	"github.com/ARMmbed/golang-utils/utils/logs"
	"github.com/ARMmbed/golang-utils/utils/parallelisation"

	"go.uber.org/atomic"
)

type logStreamer struct {
	io.Writer
	IsStdErr bool
	Loggers  logs.Loggers
}
type Subprocess struct {
	mu                    sync.RWMutex
	parentCtx             context.Context
	cancellableCtx        atomic.Value
	cancelStore           *parallelisation.CancelFunctionStore
	cmdCanceller          context.CancelFunc
	command               *exec.Cmd
	loggers               logs.Loggers
	messageOnSuccess      string
	messageOnFailure      string
	messageOnProcessStart string
	subprocess            *os.Process
	isRunning             atomic.Bool
}

func (l *logStreamer) Write(p []byte) (n int, err error) {
	text := string(p)
	if l.IsStdErr {
		l.Loggers.LogError(text)
	} else {
		l.Loggers.Log(text)
	}
	return len(p), nil
}

// Creates a subprocess description.
func New(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.Setup(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	return
}

// Executes a command
func Execute(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	p, err := New(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// In GO, there is no reentrant locks and so following what is described there
// https://groups.google.com/forum/#!msg/golang-nuts/XqW1qcuZgKg/Ui3nQkeLV80J
func (s *Subprocess) check() (err error) {
	if s.command == nil {
		err = fmt.Errorf("missing command: %w", commonerrors.ErrUndefined)
		return
	}
	if s.loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	err = s.loggers.Check()
	return
}

// Checks whether the subprocess is correctly defined.
func (s *Subprocess) Check() (err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.check()
}

func (s *Subprocess) resetContext() {
	subctx, cancelFunc := context.WithCancel(s.parentCtx)
	s.cancellableCtx.Store(subctx)
	s.cancelStore.RegisterCancelFunction(cancelFunc)
}

// Sets up a sub-process.
func (s *Subprocess) Setup(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	if s.IsOn() {
		err = s.Stop()
		if err != nil {
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning.Store(false)
	s.loggers = loggers
	s.cancelStore = parallelisation.NewCancelFunctionsStore()
	s.parentCtx = ctx
	cmdCtx, cmdcancelFunc := context.WithCancel(s.parentCtx)
	s.cmdCanceller = cmdcancelFunc
	s.command = exec.CommandContext(cmdCtx, cmd, args...)
	s.command.Stdout = &logStreamer{IsStdErr: false, Loggers: loggers}
	s.command.Stderr = &logStreamer{IsStdErr: true, Loggers: loggers}
	s.messageOnProcessStart = messageOnStart
	s.messageOnSuccess = messageOnSuccess
	s.messageOnFailure = messageOnFailure
	if s.messageOnProcessStart == "" {
		s.messageOnProcessStart = fmt.Sprintf("Executing command  -> `%v`", s.command.Path)
	}
	if s.messageOnSuccess == "" {
		s.messageOnSuccess = fmt.Sprintf("command  -> `%v` ended successfully", s.command.Path)
	}
	if s.messageOnFailure == "" {
		s.messageOnFailure = fmt.Sprintf("Error occurred when executing -> `%v`: ", s.command.Path)
	}
	return s.check()
}

// States whether the subprocess is on or not
func (s *Subprocess) IsOn() bool {
	return s.isRunning.Load()
}

// Starts the process if not already started
func (s *Subprocess) Start() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.IsOn() {
		return
	}
	s.runProcessStatusCheck()
	err = commonerrors.ConvertContextError(s.command.Start())
	if err != nil {
		s.loggers.LogError(fmt.Sprintf("Failed starting process `%v`: %v", s.command.Path, err))
		s.isRunning.Store(false)
		return
	}
	s.subprocess = s.command.Process
	s.isRunning.Store(true)
	s.loggers.Log(fmt.Sprintf("Started process [%v]", s.subprocess.Pid))
	return
}

func (s *Subprocess) Cancel() {
	store := s.cancelStore
	if store != nil {
		store.Cancel()
	}
}

func (s *Subprocess) runProcessStatusCheck() {
	s.resetContext()
	go func() {
		<-s.cancellableCtx.Load().(context.Context).Done()
		s.Cancel()
		_ = s.Stop()
	}()
}

// Executes the command and waits to completion.
func (s *Subprocess) Execute() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		s.Cancel()
	}()

	if s.IsOn() {
		return fmt.Errorf("process is already started: %w", commonerrors.ErrConflict)
	}
	s.loggers.Log(s.messageOnProcessStart)
	s.cancelStore.RegisterCancelFunction(s.cmdCanceller)
	s.runProcessStatusCheck()
	s.isRunning.Store(true)
	err = commonerrors.ConvertContextError(s.command.Run())
	s.isRunning.Store(false)
	if err == nil {
		s.loggers.Log(s.messageOnSuccess)
	} else {
		s.loggers.LogError(s.messageOnFailure, err)
	}
	return
}

// Stops the process if currently working
func (s *Subprocess) Stop() (err error) {
	if !s.IsOn() {
		return
	}
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.Cancel()
	if !s.IsOn() {
		return
	}
	s.loggers.Log(fmt.Sprintf("Stopping process [%v]", s.subprocess.Pid))
	_ = s.subprocess.Kill()
	_ = s.command.Wait()
	s.command.Process = nil
	s.subprocess = nil
	s.isRunning.Store(false)
	return
}

// Restarts a process.
func (s *Subprocess) Restart() (err error) {
	err = s.Stop()
	if err != nil {
		return
	}
	return s.Start()
}
