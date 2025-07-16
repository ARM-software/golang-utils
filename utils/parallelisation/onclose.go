package parallelisation

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type CloserStore struct {
	store[io.Closer]
}

func (s *CloserStore) RegisterCloser(closerObj ...io.Closer) {
	s.store.RegisterFunction(closerObj...)
}

func (s *CloserStore) Close() error {
	return s.Execute(context.Background())
}

func (s *CloserStore) Len() int {
	return s.store.Len()
}

// NewCloserStore returns a store of io.Closer object which will all be closed concurrently on Close(). The first error received will be returned
func NewCloserStore(stopOnFirstError bool) *CloserStore {
	return &CloserStore{
		store: *newFunctionStore[io.Closer](false, stopOnFirstError, func(_ context.Context, closerObj io.Closer) error {
			if closerObj == nil {
				return commonerrors.UndefinedVariable("closer object")
			}
			return closerObj.Close()
		}),
	}
}

// CloseAll calls concurrently Close on all io.Closer implementations passed as arguments and returns the first error encountered
func CloseAll(cs ...io.Closer) error {
	group := NewCloserStore(false)
	group.RegisterFunction(cs...)
	return group.Close()
}

// CloseAllWithContext is similar to CloseAll but can be controlled using a context.
func CloseAllWithContext(ctx context.Context, cs ...io.Closer) error {
	group := NewCloserStore(false)
	group.RegisterFunction(cs...)
	return group.Execute(ctx)
}

// CloseAllFunc calls concurrently all Close functions passed as arguments and returns the first error encountered
func CloseAllFunc(cs ...CloseFunc) error {
	group := NewCloseFunctionStoreStore(false)
	group.RegisterFunction(cs...)
	return group.Close()
}

type CloseFunc func() error

type CloseFunctionStore struct {
	store[CloseFunc]
}

func (s *CloseFunctionStore) RegisterCloseFunction(closerObj ...CloseFunc) {
	s.store.RegisterFunction(closerObj...)
}

func (s *CloseFunctionStore) Close() error {
	return s.Execute(context.Background())
}

func (s *CloseFunctionStore) Len() int {
	return s.store.Len()
}

// NewCloseFunctionStoreStore returns a store closing functions which will all be called concurrently on Close(). The first error received will be returned.
func NewCloseFunctionStoreStore(stopOnFirstError bool) *CloseFunctionStore {
	return &CloseFunctionStore{
		store: *newFunctionStore[CloseFunc](false, stopOnFirstError, func(_ context.Context, closerObj CloseFunc) error {
			return closerObj()
		}),
	}
}
