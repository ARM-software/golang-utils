package warden

import "context"

type graceful struct {
	base       IState
	neverDying chan struct{}
}

// NewGracefulWrapper returns a wrapper that ignores the dying signal.
//
// This can be useful for downstream consumers that should keep draining their
// inputs even after the upstream state has started shutting down, while still
// relying on the wrapped state for goroutine tracking and final dead/wait
// semantics.
func NewGracefulWrapper(base IState) IState {
	return &graceful{base: base, neverDying: make(chan struct{})}
}

func (g *graceful) Context(parent context.Context) context.Context {
	if parent == nil {
		return context.Background()
	}
	return parent
}

func (g *graceful) Go(f func() error) error { return g.base.Go(f) }
func (g *graceful) Kill(e error) error      { return g.base.Kill(e) }
func (g *graceful) Err() error              { return g.base.Err() }
func (g *graceful) Alive() bool             { return g.base.Alive() }
func (g *graceful) Dying() <-chan struct{}  { return g.neverDying }
func (g *graceful) Dead() <-chan struct{}   { return g.base.Dead() }
func (g *graceful) Wait() error             { return g.base.Wait() }
