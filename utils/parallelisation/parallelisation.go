package parallelisation

import (
	"context"
	"reflect"
	"time"

	"github.com/ARMmbed/build-service-common/common/commonerrors"
)

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
	completed := false

	go func() {
		channel <- blockingAction(stop)
	}()

	select {
	case err = <-channel:
		completed = true
	case <-time.After(timeout):
		stop <- true
		err = commonerrors.ErrTimeout
	}
	if !completed {
		<-channel
	}
	return
}
