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
			err = s.postStop(processErr)
			if err != nil {
				err = fmt.Errorf("%w: error running post-stop hook: %v", commonerrors.ErrUnexpected, err.Error())
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
			parallelisation.SleepWithContext(ctx, s.restartDelay)
		}

		// restart process
	}
}
