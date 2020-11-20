package subprocess

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/ARMmbed/golang-utils/utils/logs"
)

type logStreamer struct {
	io.Writer
	IsStdErr bool
	Loggers  logs.Loggers
}
type Subprocess struct {
	mu                    sync.RWMutex
	Command               *exec.Cmd
	Loggers               logs.Loggers
	MessageOnSuccess      string
	MessageOnFailure      string
	MessageOnProcessStart string
	Subprocess            *os.Process
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
func Create(loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.Setup(loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	return
}

// Executes a command
func Execute(loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	p, err := Create(loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// In GO, there is no reentrant locks and so following what is described there
// https://groups.google.com/forum/#!msg/golang-nuts/XqW1qcuZgKg/Ui3nQkeLV80J
func (s *Subprocess) check() (err error) {
	if s.Command == nil {
		err = errors.New("undefined command")
		return
	}
	if s.Loggers == nil {
		err = errors.New("undefined loggers")
		return
	}
	err = s.Loggers.Check()
	return
}

// Checks whether the subprocess is correctly defined.
func (s *Subprocess) Check() (err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.check()
}

// Sets up a sub-process.
func (s *Subprocess) Setup(loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	if s.IsOn() {
		err = s.Stop()
		if err != nil {
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Loggers = loggers
	s.Command = exec.Command(cmd, args...)
	s.Command.Stdout = &logStreamer{IsStdErr: false, Loggers: loggers}
	s.Command.Stderr = &logStreamer{IsStdErr: true, Loggers: loggers}
	s.MessageOnProcessStart = messageOnStart
	s.MessageOnSuccess = messageOnSuccess
	s.MessageOnFailure = messageOnFailure
	if s.MessageOnProcessStart == "" {
		s.MessageOnProcessStart = fmt.Sprintf("Executing command  -> `%v`", s.Command.Path)
	}
	if s.MessageOnSuccess == "" {
		s.MessageOnSuccess = fmt.Sprintf("Command  -> `%v` ended successfully", s.Command.Path)
	}
	if s.MessageOnFailure == "" {
		s.MessageOnFailure = fmt.Sprintf("Error occurred when executing -> `%v`: ", s.Command.Path)
	}
	return s.check()
}

func (s *Subprocess) isOn() bool {
	return s.Subprocess != nil
}

// States whether the subprocess is on or not
func (s *Subprocess) IsOn() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isOn()
}

// Starts the process if not already started
func (s *Subprocess) Start() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isOn() {
		return
	}
	err = s.Command.Start()
	if err != nil {
		s.Loggers.LogError(fmt.Sprintf("Failed starting process `%v`: %v", s.Command.Path, err))
		return
	}
	s.Subprocess = s.Command.Process
	s.Loggers.Log(fmt.Sprintf("Started process [%v]", s.Subprocess.Pid))
	return
}

// Executes the command and waits to completion.
func (s *Subprocess) Execute() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isOn() {
		return errors.New("process is already started")
	}
	s.Loggers.Log(s.MessageOnProcessStart)
	err = s.Command.Run()
	if err == nil {
		s.Loggers.Log(s.MessageOnSuccess)

	} else {
		s.Loggers.LogError(s.MessageOnFailure, err)
	}
	return
}

// Stops the process if currently working
func (s *Subprocess) Stop() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isOn() {
		return
	}
	s.Loggers.Log(fmt.Sprintf("Stopping process [%v]", s.Subprocess.Pid))
	_ = s.Subprocess.Kill()
	_ = s.Command.Wait()
	s.Command.Process = nil
	s.Subprocess = nil
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
