/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"fmt"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
)

// INTERNAL
// Messages logged
// Object in charge of logging subprocess output.
type subprocessMessaging struct {
	loggers               logs.Loggers
	commandPath           string
	messageOnSuccess      string
	messageOnFailure      string
	messageOnProcessStart string
	pid                   atomic.Int64
}

// Logs subprocess start.
func (s *subprocessMessaging) LogStart() {
	s.loggers.Log(s.messageOnProcessStart)
}

// Logs when subprocess failed to start.
func (s *subprocessMessaging) LogFailedStart(err error) {
	s.loggers.LogError(fmt.Sprintf("Failed starting process `%v`: %v", s.commandPath, err))
}

// Logs when subprocess has started.
func (s *subprocessMessaging) LogStarted() {
	s.loggers.Log(fmt.Sprintf("Started process [%v]", s.pid.Load()))
}

// Logs when subprocess is asked to stop.
func (s *subprocessMessaging) LogStopping() {
	s.loggers.Log(fmt.Sprintf("Stopping process [%v]", s.pid.Load()))
}

// Logs subprocess end with err if an error occurred.
func (s *subprocessMessaging) LogEnd(err error) {
	if err == nil {
		s.loggers.Log(s.messageOnSuccess)
	} else {
		s.loggers.LogError(s.messageOnFailure, err)
	}
}

// Sets the process PID.
func (s *subprocessMessaging) SetPid(pid int) {
	s.pid.Store(int64(pid))
}

func (s *subprocessMessaging) Check() (err error) {
	if s.loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	err = s.loggers.Check()
	return
}

func newSubprocessMessaging(loggers logs.Loggers, messageOnSuccess string, messageOnFailure string, messageOnProcessStart string, commandPath string) *subprocessMessaging {
	m := &subprocessMessaging{
		loggers:               loggers,
		commandPath:           commandPath,
		messageOnSuccess:      messageOnSuccess,
		messageOnFailure:      messageOnFailure,
		messageOnProcessStart: messageOnProcessStart,
	}
	m.messageOnFailure = messageOnFailure
	if m.messageOnProcessStart == "" {
		m.messageOnProcessStart = fmt.Sprintf("Executing command  -> `%v`", commandPath)
	}
	if m.messageOnSuccess == "" {
		m.messageOnSuccess = fmt.Sprintf("command  -> `%v` ended successfully", commandPath)
	}
	if m.messageOnFailure == "" {
		m.messageOnFailure = fmt.Sprintf("Error occurred when executing -> `%v`: ", commandPath)
	}
	return m
}
