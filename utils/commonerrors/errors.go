package commonerrors

import (
	"context"
	"errors"
)

var (
	ErrNotImplemented     = errors.New("not implemented")
	ErrNoExtension        = errors.New("missing extension")
	ErrNoLogger           = errors.New("missing logger")
	ErrNoLoggerSource     = errors.New("missing logger source")
	ErrNoLogSource        = errors.New("missing log source")
	ErrUndefined          = errors.New("undefined")
	ErrInvalidDestination = errors.New("invalid destination")
	ErrTimeout            = errors.New("timeout")
	ErrLocked             = errors.New("locked")
	ErrNotFound           = errors.New("not found")
	ErrUnsupported        = errors.New("unsupported")
	ErrUnavailable        = errors.New("unavailable")
	ErrWrongUser          = errors.New("wrong user")
	ErrUnknown            = errors.New("unknown")
	ErrInvalid            = errors.New("invalid")
	ErrConflict           = errors.New("conflict")
	ErrMarshalling        = errors.New("unserialisable")
	ErrCancelled          = errors.New("cancelled")
)

func Any(target error, err ...error) bool {
	for _, e := range err {
		if errors.Is(e, target) || errors.Is(target, e) {
			return true
		}
	}
	return false
}

func None(target error, err ...error) bool {
	for _, e := range err {
		if errors.Is(e, target) || errors.Is(target, e) {
			return false
		}
	}
	return true
}

// Determines what the context error is if any.
func DetermineContextError(ctx context.Context) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}
	if Any(err, context.Canceled) {
		return ErrCancelled
	}
	if Any(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return err
}
