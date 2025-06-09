package parallelisation

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// IErrorWaiter can be used to wait on errgroups and similar types where Wait() returns an error
// This is used to support use in the WaitWithContextAndError function to wait but listen to contexts
type IErrorWaiter interface {
	Wait() error
}

func WaitWithContextAndError(ctx context.Context, wg IErrorWaiter) (err error) {
	done := make(chan struct{})
	var g errgroup.Group
	g.SetLimit(1)
	g.Go(func() error {
		defer close(done)
		return wg.Wait()
	})
	select {
	case <-ctx.Done():
		return DetermineContextError(ctx)
	case <-done:
		return g.Wait() // since there is only one this will return when wg does
	}
}

// IError can be used to wait on sync WaitGroups and similar types where Wait() does not return an error
// This is used to support use in the WaitWithContext function to wait but listen to contexts
type IWaiter interface {
	Wait()
}

func WaitWithContext(ctx context.Context, wg IWaiter) (err error) {
	done := make(chan struct{})
	var g errgroup.Group
	g.SetLimit(1)
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-ctx.Done():
		return DetermineContextError(ctx)
	case <-done:
		return nil
	}
}
