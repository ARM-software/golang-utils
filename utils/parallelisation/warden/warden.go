package warden

import (
	"context"
	"sync"

	"github.com/sasha-s/go-deadlock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

var (
	// ErrStillAlive indicates the state has not started shutting down.
	ErrStillAlive = commonerrors.New(commonerrors.ErrConflict, "warden: still alive")
	// ErrDying is used internally to indicate the state is already dying.
	ErrDying = commonerrors.New(commonerrors.ErrInterrupted, "warden: dying")
	// ErrSpawnAfterDeath indicates Go was called after all tracked goroutines terminated.
	ErrSpawnAfterDeath = commonerrors.New(commonerrors.ErrConflict, "warden: Go called after all goroutines terminated")
	// ErrKillWhileAlive indicates Kill was called with ErrStillAlive.
	ErrKillWhileAlive = commonerrors.New(commonerrors.ErrConflict, "warden: Kill with ErrStillAlive")
)

type stateContext struct {
	context     context.Context
	cancelStore *parallelisation.CancelFunctionStore
	done        <-chan struct{}
}

func newStateContext(ctx context.Context) *stateContext {
	return &stateContext{
		context:     ctx,
		cancelStore: parallelisation.NewCancelFunctionsStore(parallelisation.ExecuteAll),
		done:        ctx.Done(),
	}
}

func newStateContextWithCancel(ctx context.Context, cancel context.CancelFunc) *stateContext {
	state := newStateContext(ctx)
	state.registerCancelFunc(cancel)
	return state
}

func (s *stateContext) registerCancelFunc(cancelFunc context.CancelFunc) {
	if s == nil || cancelFunc == nil {
		return
	}
	s.cancelStore.RegisterCancelFunction(cancelFunc)
}

func (s *stateContext) Cancel() {
	if s == nil || s.cancelStore == nil {
		return
	}
	s.cancelStore.Cancel()
}

// Warden tracks the lifecycle of one or more goroutines.
type Warden struct {
	mu         deadlock.Mutex
	initOnce   sync.Once
	aliveCount int
	dying      chan struct{}
	dead       chan struct{}
	reason     error
	contexts   map[context.Context]*stateContext
}

// New returns a new live warden.
func New() *Warden {
	var w Warden
	w.init()
	return &w
}

// WithContext returns a new warden that is killed when parent is cancelled.
func WithContext(parent context.Context) IState {
	w := New()
	if parent != nil && parent.Done() != nil {
		go func() {
			select {
			case <-w.Dying():
			case <-parent.Done():
				_ = w.Kill(parent.Err())
			}
		}()
	}
	wCtx, cancel := context.WithCancel(parentOrBackground(parent))
	w.addContext(parentOrBackground(parent), newStateContextWithCancel(wCtx, cancel))
	return w
}

// Context returns a child context of parent that is cancelled when the warden
// starts dying or when parent is cancelled.
func (w *Warden) Context(parent context.Context) context.Context {
	w.init()
	parent = parentOrBackground(parent)

	w.mu.Lock()
	defer w.mu.Unlock()

	if state, ok := w.contexts[parent]; ok {
		return state.context
	}

	child, cancel := context.WithCancel(parent)
	w.addContext(parent, newStateContextWithCancel(child, cancel))
	return child
}

func (w *Warden) addContext(parent context.Context, child *stateContext) {
	if !commonerrors.Any(w.reason, ErrStillAlive) {
		child.Cancel()
		return
	}

	w.contexts[parent] = child
	for parentCtx, childCtx := range w.contexts {
		select {
		case <-childCtx.done:
			delete(w.contexts, parentCtx)
		default:
		}
	}
}

func (w *Warden) init() {
	w.initOnce.Do(func() {
		w.contexts = make(map[context.Context]*stateContext)
		w.dying = make(chan struct{})
		w.dead = make(chan struct{})
		w.reason = ErrStillAlive
	})
}

// Dead returns a channel that closes when all tracked goroutines have terminated.
func (w *Warden) Dead() <-chan struct{} {
	w.init()
	return w.dead
}

// Dying returns a channel that closes when the warden starts shutting down.
func (w *Warden) Dying() <-chan struct{} {
	w.init()
	return w.dying
}

// Wait blocks until all tracked goroutines have returned and then returns the death reason.
func (w *Warden) Wait() error {
	w.init()
	<-w.dead
	w.mu.Lock()
	reason := w.reason
	w.mu.Unlock()
	return reason
}

// Go runs f in a new tracked goroutine.
func (w *Warden) Go(f func() error) error {
	w.init()
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.dead:
		return ErrSpawnAfterDeath
	default:
	}

	w.aliveCount++
	go w.run(f)
	return nil
}

func (w *Warden) run(f func() error) {
	err := f()
	w.mu.Lock()
	defer w.mu.Unlock()

	w.aliveCount--
	if w.aliveCount == 0 || err != nil {
		_ = w.kill(err)
		if w.aliveCount == 0 {
			close(w.dead)
		}
	}
}

// Kill moves the warden into dying mode for the supplied reason.
func (w *Warden) Kill(reason error) error {
	w.init()
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.kill(reason)
}

func (w *Warden) kill(reason error) (err error) {
	if commonerrors.Any(reason, ErrStillAlive) {
		return ErrKillWhileAlive
	}
	if commonerrors.Any(reason, ErrDying) {
		if commonerrors.Any(w.reason, ErrStillAlive) {
			return commonerrors.New(commonerrors.ErrConflict, "warden: Kill with ErrDying while still alive")
		}
		return nil
	}
	if commonerrors.Any(w.reason, ErrStillAlive) {
		w.reason = reason
		close(w.dying)
		for _, child := range w.contexts {
			child.Cancel()
		}
		clear(w.contexts)
		return nil
	}
	if commonerrors.Any(w.reason, nil) {
		w.reason = reason
	}
	return nil
}

// Err returns the death reason, or ErrStillAlive while the warden is alive.
func (w *Warden) Err() (reason error) {
	w.init()
	w.mu.Lock()
	reason = w.reason
	w.mu.Unlock()
	return reason
}

// Alive reports whether the warden has not started dying.
func (w *Warden) Alive() bool {
	return commonerrors.Any(w.Err(), ErrStillAlive)
}

func parentOrBackground(parent context.Context) context.Context {
	if parent == nil {
		return context.Background()
	}
	return parent
}
