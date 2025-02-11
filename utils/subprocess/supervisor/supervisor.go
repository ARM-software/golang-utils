package supervisor

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/subprocess"
)

// A Supervisor will run a command and automatically restart it if it exits. Hooks can be used to execute code at
// different points in the execution lifecyle. Restarts can be delayed
type Supervisor struct {
	newCommand    func(ctx context.Context) *subprocess.Subprocess
	preStart      func(context.Context) error
	postStart     func(context.Context) error
	postStop      func(context.Context, error) error
	haltingErrors []error
	restartDelay  time.Duration
}

type SupervisorOption func(*Supervisor)

// NewSupervisor will take a function 'newCommand' that creates a command and run it every time the command exits.
// Options can be supplied by the 'opts' variadic argument that control the lifecyle of the supervisor
func NewSupervisor(newCommand func(ctx context.Context) *subprocess.Subprocess, opts ...SupervisorOption) *Supervisor {
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
func WithPostStop(function func(context.Context, error) error) SupervisorOption {
	return func(s *Supervisor) {
		s.postStop = function
	}
}

// WithHaltingErrors are errors that won't trigger the supervisor to restart
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

// Run will run the supervisor and execute any of the command hooks. If it recieves a halting error or the context is cancelled then it will exit
func (s *Supervisor) Run(ctx context.Context) (err error) {
	for {
		err = parallelisation.DetermineContextError(ctx)
		if err != nil {
			return
		}

		if s.preStart != nil {
			err = s.preStart(ctx)
			if err != nil {
				err = fmt.Errorf("%w: error running pre-start hook: %v", commonerrors.ErrUnexpected, err.Error())
				return
			}
		}

		g, _ := errgroup.WithContext(ctx)
		cmd := s.newCommand(ctx)
		g.Go(cmd.Execute)

		if s.postStart != nil {
			err = s.postStart(ctx)
			if err != nil {
				err = fmt.Errorf("%w: error running post-start hook: %v", commonerrors.ErrUnexpected, err.Error())
				return err
			}
		}

		processErr := g.Wait()

		if s.postStop != nil {
			err = s.postStop(context.WithoutCancel(ctx), processErr)
			if err != nil {
				err = fmt.Errorf("%w: error running post-stop hook: %v", commonerrors.ErrUnexpected, err.Error())
				return err
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
}
