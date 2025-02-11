package supervisor

import (
	"context"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/subprocess"
)

type ISupervisor interface {
	Run(ctx context.Context) error
}

type Supervisor struct {
	newCommand   func(ctx context.Context) *subprocess.Subprocess
	preStart     func(context.Context) error
	postStart    func(context.Context) error
	postStop     func(error) error
	ignoreErrors []error
	restartDelay time.Duration
}

type SupervisorOption func(*Supervisor)

func NewSupervisor(cmdFactory func(ctx context.Context) *subprocess.Subprocess, opts ...SupervisorOption) *Supervisor {
	supervisor := &Supervisor{
		newCommand:   cmdFactory,
		restartDelay: 0,
	}
	for _, opt := range opts {
		opt(supervisor)
	}
	return supervisor
}

func WithPreStart(function func(context.Context) error) SupervisorOption {
	return func(s *Supervisor) {
		s.preStart = function
	}
}

func WithPostStart(function func(context.Context) error) SupervisorOption {
	return func(s *Supervisor) {
		s.postStart = function
	}
}

func WithPostStop(function func(error) error) SupervisorOption {
	return func(s *Supervisor) {
		s.postStop = function
	}
}

func WithIgnoreErrors(errs ...error) SupervisorOption {
	return func(s *Supervisor) {
		s.ignoreErrors = errs
	}
}

func WithRestartDelay(delay time.Duration) SupervisorOption {
	return func(s *Supervisor) {
		s.restartDelay = delay
	}
}

func (s *Supervisor) Run(ctx context.Context) (err error) {
	for {
		if s.preStart != nil {
			err = s.preStart(ctx)
			if err != nil {
				return
			}
		}

		cmd := s.newCommand(ctx)

		done := make(chan error, 1)
		go func() {
			done <- cmd.Execute()
		}()

		if s.postStart != nil {
			err = s.postStart(ctx)
			if err != nil {
				return err
			}
		}

		var processErr error
		select {
		case <-ctx.Done():
			if cmd.IsOn() {
				cmd.Cancel()
			}
			return commonerrors.ConvertContextError(ctx.Err())
		case processErr = <-done:
			// done
		}

		if s.postStop != nil {
			err = s.postStop(processErr)
			if err != nil {
				return err
			}
		}

		if processErr != nil {
			if commonerrors.Any(processErr, s.ignoreErrors...) ||
				commonerrors.RelatesTo(processErr.Error(), s.ignoreErrors...) {
				return processErr
			}
		}

		if s.restartDelay > 0 {
			select {
			case <-time.After(s.restartDelay):
			case <-ctx.Done():
				return commonerrors.ConvertContextError(ctx.Err())
			}
		}

		// restart process
	}
}
