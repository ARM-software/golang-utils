package parallelisation

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type IWaiter interface {
	Wait() error
}

func WaitWithContext(ctx context.Context, wg IWaiter) (err error) {
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
