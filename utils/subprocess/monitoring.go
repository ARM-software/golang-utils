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

// Interrupts an on-going process.
func (s *subprocessMonitoring) CancelSubprocess() {
	s.monitoringStopping.Store(true)
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
