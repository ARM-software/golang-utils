package commonerrors

import "errors"

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
)
