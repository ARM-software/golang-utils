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
	option := ExecuteAll
	if stopOnFirstError {
		option = StopOnFirstError
	}
	return NewCloserStoreWithOptions(option, Parallel)
}

// NewCloserStoreWithOptions returns a store of io.Closer object which will all be closed on Close(). The first error received if any will be returned
func NewCloserStoreWithOptions(opts ...StoreOption) *CloserStore {
	return &CloserStore{
		store: *newFunctionStore[io.Closer](func(_ context.Context, closerObj io.Closer) error {
			if closerObj == nil {
				return commonerrors.UndefinedVariable("closer object")
			}
			return closerObj.Close()
		}, append(opts, RetainAfterExecution)...),
	}
}

// CloseAll calls concurrently Close on all io.Closer implementations passed as arguments and returns the first error encountered
func CloseAll(cs ...io.Closer) error {
	group := NewCloserStore(false)
	group.RegisterFunction(cs...)
	return group.Close()
}

// CloseAllAndCollateErrors calls concurrently Close on all io.Closer implementations passed as arguments and returns the errors encountered
func CloseAllAndCollateErrors(cs ...io.Closer) error {
	group := NewCloserStoreWithOptions(ExecuteAll, Parallel, JoinErrors)
	group.RegisterFunction(cs...)
	return group.Close()
}

// CloseAllWithContext is similar to CloseAll but can be controlled using a context.
func CloseAllWithContext(ctx context.Context, cs ...io.Closer) error {
	group := NewCloserStore(false)
	group.RegisterFunction(cs...)
	return group.Execute(ctx)
}

// CloseAllWithContextAndCollateErrors is similar to CloseAllAndCollateErrors but can be controlled using a context.
func CloseAllWithContextAndCollateErrors(ctx context.Context, cs ...io.Closer) error {
	group := NewCloserStoreWithOptions(ExecuteAll, Parallel, JoinErrors)
	group.RegisterFunction(cs...)
	return group.Execute(ctx)
}

// CloseAllFunc calls concurrently all Close functions passed as arguments and returns the first error encountered
func CloseAllFunc(cs ...CloseFunc) error {
	group := NewCloseFunctionStoreStore(false)
	group.RegisterFunction(cs...)
	return group.Close()
}

// CloseAllFuncAndCollateErrors calls concurrently all Close functions passed as arguments and returns the errors encountered
func CloseAllFuncAndCollateErrors(cs ...CloseFunc) error {
	group := NewCloseFunctionStore(ExecuteAll, Parallel, JoinErrors)
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

func (s *CloseFunctionStore) RegisterCancelStore(cancelStore *CancelFunctionStore) {
	if cancelStore == nil {
		return
	}
	s.store.RegisterFunction(func() error {
		cancelStore.Cancel()
		return nil
	})
}

func (s *CloseFunctionStore) RegisterCancelFunction(cancelFunc ...context.CancelFunc) {
	cancelStore := NewCancelFunctionsStore()
	cancelStore.RegisterCancelFunction(cancelFunc...)
	s.RegisterCancelStore(cancelStore)
}

func (s *CloseFunctionStore) Close() error {
	return s.Execute(context.Background())
}

func (s *CloseFunctionStore) Len() int {
	return s.store.Len()
}

// NewCloseFunctionStore returns a store closing functions which will all be called on Close(). The first error received if any will be returned.
func NewCloseFunctionStore(options ...StoreOption) *CloseFunctionStore {
	return &CloseFunctionStore{
		store: *newFunctionStore[CloseFunc](func(_ context.Context, closerObj CloseFunc) error {
			return closerObj()
		}, append(options, RetainAfterExecution)...),
	}
}

// NewCloseFunctionStoreStore is exactly the same as NewConcurrentCloseFunctionStore but without a typo in the name.
func NewCloseFunctionStoreStore(stopOnFirstError bool) *CloseFunctionStore {
	return NewConcurrentCloseFunctionStore(stopOnFirstError)
}

// NewConcurrentCloseFunctionStore returns a store closing functions which will all be called concurrently on Close(). The first error received will be returned.
// Prefer using NewCloseFunctionStore where possible
func NewConcurrentCloseFunctionStore(stopOnFirstError bool) *CloseFunctionStore {
	option := ExecuteAll
	if stopOnFirstError {
		option = StopOnFirstError
	}
	return NewCloseFunctionStore(option, Parallel)
}
