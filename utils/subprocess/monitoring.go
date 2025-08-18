/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"context"
	"time"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// INTERNAL
// Object in charge of monitoring subprocesses.
type subprocessMonitoring struct {
	cancelStore        *parallelisation.CancelFunctionStore
	parentCtx          context.Context
	cancellableCtx     atomic.Value
	monitoringOn       atomic.Bool
	monitoringStopping atomic.Bool
}

func newSubprocessMonitoring(parentCtx context.Context) *subprocessMonitoring {
	m := &subprocessMonitoring{
		cancelStore: parallelisation.NewCancelFunctionsStore(),
		parentCtx:   parentCtx,
	}
	m.Reset()
	return m
}

// CancelSubprocess interrupts an on-going process.
func (s *subprocessMonitoring) CancelSubprocess() {
	// Ensure we only ever run the cancel-store once and prevent the following deadlocks:
	// 1. Some functions like Execute() do defer s.Cancel()
	// 2. This calls subprocessMonitoring.CancelSubprocess() which calls cancelStore.Cancel() (acquiring mutex)
	// 3. That Cancel() calls the context cancel func, which closes ProcessContext().Done()
	// 4. The runProcessMonitoring blocks on ctx<-Done() and calls m.CancelSubprocess() again
	// 5. This tries to run cancelStore.Cancel() a second time while the first Cancel() is still executing
	// 6. go-deadlock detects deadlock
	if s.monitoringStopping.Swap(true) {
		return
	}
	s.cancelStore.Cancel()
}

func (s *subprocessMonitoring) RunMonitoring(stopProcess func() error) {
	timeoutCtx, cancel := context.WithTimeout(s.parentCtx, time.Second)
	defer cancel()
	for s.IsOn() {
		// if monitoring is already on, either wait until it is stopped to relaunch it or do not do anything
		if !s.monitoringStopping.Load() {
			return
		}
		parallelisation.SleepWithContext(timeoutCtx, time.Millisecond)
		err := parallelisation.DetermineContextError(timeoutCtx)
		if err != nil {
			return
		}
	}
	s.Reset()
	s.runProcessMonitoring(stopProcess)
}

func (s *subprocessMonitoring) Reset() {
	s.monitoringStopping.Store(false)
	subctx, cancelFunc := context.WithCancel(s.parentCtx)
	s.cancellableCtx.Store(subctx)
	s.cancelStore.RegisterCancelFunction(cancelFunc)
}

func (s *subprocessMonitoring) IsOn() bool {
	return s.monitoringOn.Load()
}

func (s *subprocessMonitoring) ProcessContext() context.Context {
	return s.cancellableCtx.Load().(context.Context)
}

func (s *subprocessMonitoring) runProcessMonitoring(stopProcess func() error) {
	go func(m *subprocessMonitoring, stop func() error) {
		m.monitoringOn.Store(true)
		<-s.ProcessContext().Done()
		m.CancelSubprocess()
		_ = stop()
		m.monitoringOn.Store(false)
	}(s, stopProcess)
}
