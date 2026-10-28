package warden

import "context"

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/parallelisation/$GOPACKAGE IState

// IState describes the lifecycle state of a set of tracked goroutines.
//
// A state starts alive, becomes dying when [IState.Kill] is called or when one
// of its tracked goroutines fails and becomes dead once all tracked goroutines
// have returned.
type IState interface {
	// Context returns a child context of parent that is cancelled when either the
	// parent context is cancelled or the state starts dying.
	Context(parent context.Context) context.Context
	// Go starts f in a new tracked goroutine.
	Go(f func() error) error
	// Kill moves the state into dying mode with the supplied reason.
	Kill(error) error
	// Err returns the death reason, or ErrStillAlive while the state is alive.
	Err() error
	// Alive reports whether the state has not started dying yet.
	Alive() bool
	// Dying is closed when the state starts shutting down.
	Dying() <-chan struct{}
	// Dead is closed when all tracked goroutines have terminated.
	Dead() <-chan struct{}
	// Wait blocks until the state is dead and then returns the death reason.
	Wait() error
}
