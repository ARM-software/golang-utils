package parallelisation

import (
	"context"
	"reflect"
	"time"

	"go.uber.org/atomic"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
)

// Determines what the context error is if any.
func DetermineContextError(ctx context.Context) error {
	return commonerrors.ConvertContextError(ctx.Err())
}

type result struct {
	Item interface{}
	err  error
}

// Parallelises an action over as many goroutines as specified by the argList and retrieves all the results when all the goroutines are done.
func Parallelise(argList interface{}, action func(arg interface{}) (interface{}, error), resultType reflect.Type) (results interface{}, err error) {
	keepReturn := resultType != nil
	argListValue := reflect.ValueOf(argList)
	len := argListValue.Len()
	channel := make(chan result, len)
	for i := 0; i < len; i++ {
		go func(args reflect.Value) {
			var r result
			r.Item, r.err = func(v reflect.Value) (interface{}, error) {
				return action(v.Interface())
			}(args)
			channel <- r
		}(argListValue.Index(i))
	}
	var v reflect.Value
	if keepReturn {
		v = reflect.MakeSlice(resultType, 0, len)
	}
	for i := 0; i < len; i++ {
		r := <-channel
		err = r.err
		if err != nil {
			return
		}
		if keepReturn {
			v = reflect.Append(v, reflect.ValueOf(r.Item))
		}
	}
	if keepReturn {
		results = v.Interface()
	}
	return
}

// Calls function `f` with a `period` and an `offset`.
func Schedule(ctx context.Context, period time.Duration, offset time.Duration, f func(time.Time)) {
	go func() {
		// Position the first execution
		first := time.Now().Truncate(period).Add(offset)
		if first.Before(time.Now()) {
			first = first.Add(period)
		}
		firstC := time.After(time.Until(first))

		// Receiving from a nil channel blocks forever
		t := &time.Ticker{C: nil}

		for {
			select {
			case v := <-firstC:
				// The ticker has to be started before f as it can take some time to finish
				t = time.NewTicker(period)
				f(v)
			case v := <-t.C:
				f(v)
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}

// Runs an action with timeout
func RunActionWithTimeout(blockingAction func(stop chan bool) error, timeout time.Duration) (err error) {
	channel := make(chan error, 1)
	stop := make(chan bool)
	completed := atomic.NewBool(false)

	go func() {
		channel <- blockingAction(stop)
	}()

	select {
	case err = <-channel:
		completed.Store(true)
	case <-time.After(timeout):
		stop <- true
		err = commonerrors.ErrTimeout
	}
	if !completed.Load() {
		<-channel
	}
	return
}

// Runs an action with timeout
// blockingAction's context will be cancelled on exit.
func RunActionWithTimeoutAndContext(ctx context.Context, timeout time.Duration, blockingAction func(context.Context) error) error {
	store := NewCancelFunctionsStore()
	defer store.Cancel()
	return RunActionWithTimeoutAndCancelStore(ctx, timeout, store, blockingAction)
}

// Runs an action with timeout
func RunActionWithTimeoutAndCancelStore(ctx context.Context, timeout time.Duration, store *CancelFunctionStore, blockingAction func(context.Context) error) error {
	timeoutContext, cancel := context.WithTimeout(ctx, timeout)
	store.RegisterCancelFunction(cancel)
	channel := make(chan error, 1)
	go func() {
		channel <- blockingAction(timeoutContext)
	}()

	err := <-channel
	err2 := DetermineContextError(timeoutContext)

	if err2 != nil {
		return err2
	}
	return err
}

// Runs an action with a check in parallel
// The function performing the check should return true if the check was favorable; false otherwise. If the check did not have the expected result and the whole function would be cancelled.
func RunActionWithParallelCheck(ctx context.Context, action func(ctx context.Context) error, checkAction func(ctx context.Context) bool, checkPeriod time.Duration) error {
	cancelStore := NewCancelFunctionsStore()
	defer cancelStore.Cancel()
	cancellableCtx, cancelFunc := context.WithCancel(ctx)
	cancelStore.RegisterCancelFunction(cancelFunc)
	go func(ctx context.Context, store *CancelFunctionStore) {
		for {
			select {
			case <-ctx.Done():
				store.Cancel()
				return
			default:
				if !checkAction(ctx) {
					store.Cancel()
					return
				}
				time.Sleep(checkPeriod)
			}
		}
	}(cancellableCtx, cancelStore)
	err := action(cancellableCtx)
	err2 := DetermineContextError(cancellableCtx)
	if err2 != nil {
		return err2
	}
	return err
}
