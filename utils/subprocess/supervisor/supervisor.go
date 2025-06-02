package supervisor

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/safecast"
	"github.com/ARM-software/golang-utils/utils/subprocess"
)

// A Supervisor will run a command and automatically restart it if it exits. Hooks can be used to execute code at
// different points in the execution lifecyle. Restarts can be delayed
type Supervisor struct {
	newCommand    func(ctx context.Context) (*subprocess.Subprocess, error)
	preStart      func(context.Context) error
	postStart     func(context.Context) error
	postStop      func(context.Context, error) error
	postEnd       func()
	haltingErrors []error
	restartDelay  time.Duration
	count         uint
	cmd           *subprocess.Subprocess
}

type SupervisorOption func(*Supervisor)

// NewSupervisor will take a function 'newCommand' that creates a command and run it every time the command exits.
// Options can be supplied by the 'opts' variadic argument that control the lifecyle of the supervisor
func NewSupervisor(newCommand func(ctx context.Context) (*subprocess.Subprocess, error), opts ...SupervisorOption) *Supervisor {
	supervisor := &Supervisor{
		newCommand:   newCommand,
		restartDelay: 0,
	}
	for _, opt := range opts {
		opt(supervisor)
	}
	return supervisor
}

// WithPreStart will run 'function' before the supervisor starts
func WithPreStart(function func(context.Context) error) SupervisorOption {
	return func(s *Supervisor) {
		s.preStart = function
	}
}

// WithPostStart will run 'function' after the supervisor has started
func WithPostStart(function func(context.Context) error) SupervisorOption {
	return func(s *Supervisor) {
		s.postStart = function
	}
}

// WithPostStop will run 'function' after the supervised command has stopped
// It's context will ignore cancellations so any timeouts should be added within
// the function body itself
func WithPostStop(function func(context.Context, error) error) SupervisorOption {
	return func(s *Supervisor) {
		s.postStop = function
	}
}

// WithHaltingErrors are errors that won't trigger the supervisor to restart and on which, the subprocess will just halt.
func WithHaltingErrors(errs ...error) SupervisorOption {
	return func(s *Supervisor) {
		s.haltingErrors = errs
	}
}

// WithRestartDelay will delay the supervisor from restarting for a period of time specified by 'delay'
func WithRestartDelay(delay time.Duration) SupervisorOption {
	return func(s *Supervisor) {
		s.restartDelay = delay
	}
}

// WithCount will run cause the supervisor to exit after 'count' executions.
func WithCount[I safecast.INumber](count I) SupervisorOption {
	return func(s *Supervisor) {
		s.count = safecast.ToUint(count)
	}
}

// WithPostEnd will run 'function' after the supervisor has stopped.
// It does not take a context to ensure that it runs after a context has been cancelled.
// It does not return an error as this could cause confusion with the other returned errors.
func WithPostEnd(function func()) SupervisorOption {
	return func(s *Supervisor) {
		s.postEnd = function
	}
}

// Run will run the supervisor and execute any of the command hooks. If it receives a halting error or the context is cancelled then it will exit
func (s *Supervisor) Run(ctx context.Context) (err error) {
	if s.postEnd != nil {
		defer s.postEnd()
	}

	for i := uint(0); s.count == 0 || i < s.count; i++ {
		err = parallelisation.DetermineContextError(ctx)
		if err != nil {
			return
		}

		if s.preStart != nil {
			err = s.preStart(ctx)
			if err != nil {
				if commonerrors.Any(err, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
					return err
				}
				return fmt.Errorf("%w: error running pre-start hook: %v", commonerrors.ErrUnexpected, err.Error())
			}
		}

		g, _ := errgroup.WithContext(ctx)
		s.cmd, err = s.newCommand(ctx)
		if err != nil {
			if commonerrors.Any(err, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
				return err
			}
			return fmt.Errorf("%w: error occurred when creating new command: %v", commonerrors.ErrUnexpected, err.Error())
		}
		if s.cmd == nil {
			return fmt.Errorf("%w: command was undefined", commonerrors.ErrUndefined)
		}

		g.Go(s.cmd.Execute)

		if s.postStart != nil {
			err = s.postStart(ctx)
			if err != nil {
				if commonerrors.Any(err, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
					return err
				}
				return fmt.Errorf("%w: error running post-start hook: %v", commonerrors.ErrUnexpected, err.Error())
			}
		}

		processErr := g.Wait()

		if s.postStop != nil {
			err = s.postStop(context.WithoutCancel(ctx), processErr)
			if err != nil {
				if commonerrors.Any(err, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
					return err
				}
				return fmt.Errorf("%w: error running post-stop hook: %v", commonerrors.ErrUnexpected, err.Error())
			}
		}

		if processErr != nil {
			if commonerrors.Any(processErr, s.haltingErrors...) ||
				commonerrors.RelatesTo(processErr.Error(), s.haltingErrors...) {
				return processErr
			}
		}

		if s.restartDelay > 0 {
			parallelisation.SleepWithContext(ctx, s.restartDelay)
		}

		// restart process
	}

	return
}
