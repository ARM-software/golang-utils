/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package commonerrors defines typical errors which can happen.
package commonerrors

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// List of common errors used to qualify and categorise go errors
// Note: if adding error types to this list, ensure mapping functions (below) are also updated.
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
	ErrStaleLock          = errors.New("stale lock")
	ErrExists             = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrUnsupported        = errors.New("unsupported")
	ErrUnavailable        = errors.New("unavailable")
	ErrWrongUser          = errors.New("wrong user")
	ErrUnauthorised       = errors.New("unauthorised")
	ErrUnknown            = errors.New("unknown")
	ErrInvalid            = errors.New("invalid")
	ErrConflict           = errors.New("conflict")
	ErrMarshalling        = errors.New("unserialisable")
	ErrCancelled          = errors.New("cancelled")
	ErrEmpty              = errors.New("empty")
	ErrUnexpected         = errors.New("unexpected")
	ErrTooLarge           = errors.New("too large")
	ErrForbidden          = errors.New("forbidden")
	ErrCondition          = errors.New("failed condition")
	ErrEOF                = errors.New("end of file")
	ErrMalicious          = errors.New("suspected malicious intent")
	ErrOutOfRange         = errors.New("out of range")
	// ErrWarning is a generic error that can be used when an error should be raised but it shouldn't necessary be
	// passed up the chain, for example in cases where an error should be logged but the program should continue. In
	// these situations it should be handled immediately and then ignored/set to nil.
	ErrWarning = errors.New(warningStr)
)

const warningStr = "warning"

var warningStrPrepend = fmt.Sprintf("%v: ", warningStr)

// IsCommonError returns whether an error is a commonerror
func IsCommonError(target error) bool {
	return Any(target, ErrNotImplemented, ErrNoExtension, ErrNoLogger, ErrNoLoggerSource, ErrNoLogSource, ErrUndefined, ErrInvalidDestination, ErrTimeout, ErrLocked, ErrStaleLock, ErrExists, ErrNotFound, ErrUnsupported, ErrUnavailable, ErrWrongUser, ErrUnauthorised, ErrUnknown, ErrInvalid, ErrConflict, ErrMarshalling, ErrCancelled, ErrEmpty, ErrUnexpected, ErrTooLarge, ErrForbidden, ErrCondition, ErrEOF, ErrMalicious, ErrWarning, ErrOutOfRange)
}

// Any determines whether the target error is of the same type as any of the errors `err`
func Any(target error, err ...error) bool {
	for i := range err {
		e := err[i]
		if errors.Is(e, target) || errors.Is(target, e) {
			return true
		}
	}
	return false
}

// None determines whether the target error is of none of the types of the errors `err`
func None(target error, err ...error) bool {
	for i := range err {
		e := err[i]
		if errors.Is(e, target) || errors.Is(target, e) {
			return false
		}
	}
	return true
}

// CorrespondTo determines whether a `target` error corresponds to a specific error described by `description`
// It will check whether the error contains the string in its description. It is not case-sensitive.
// ```code
//
//	  CorrespondTo(errors.New("feature a is not supported"), "not supported") = True
//	```
func CorrespondTo(target error, description ...string) bool {
	if target == nil {
		return false
	}
	desc := strings.ToLower(target.Error())
	for i := range description {
		d := strings.ToLower(description[i])
		if desc == d || strings.Contains(desc, d) {
			return true
		}
	}
	return false
}

// RelatesTo determines whether an error description string could relate to a particular set of common errors
// This assumes that the error description follows the convention of placing the type of errors at the start of the string.
func RelatesTo(target string, errors ...error) bool {
	for i := range errors {
		if strings.HasPrefix(target, errors[i].Error()) {
			return true
		}
	}
	return false
}

// deserialiseCommonError returns the common error corresponding to its string value
func deserialiseCommonError(errStr string) (bool, error) {
	errStr = strings.TrimSpace(errStr)
	switch {
	case errStr == "":
		return true, nil
	case errStr == ErrInvalid.Error():
		return true, ErrInvalid
	case errStr == ErrNotFound.Error():
		return true, ErrNotFound
	case CorrespondTo(ErrNotImplemented, errStr):
		return true, ErrNotImplemented
	case CorrespondTo(ErrNoExtension, errStr):
		return true, ErrNoExtension
	case CorrespondTo(ErrNoLogger, errStr):
		return true, ErrNoLogger
	case CorrespondTo(ErrNoLoggerSource, errStr):
		return true, ErrNoLoggerSource
	case CorrespondTo(ErrNoLogSource, errStr):
		return true, ErrNoLogSource
	case CorrespondTo(ErrUndefined, errStr):
		return true, ErrUndefined
	case CorrespondTo(ErrInvalidDestination, errStr):
		return true, ErrInvalidDestination
	case CorrespondTo(ErrTimeout, errStr):
		return true, ErrTimeout
	case CorrespondTo(ErrLocked, errStr):
		return true, ErrLocked
	case CorrespondTo(ErrStaleLock, errStr):
		return true, ErrStaleLock
	case CorrespondTo(ErrExists, errStr):
		return true, ErrExists
	case CorrespondTo(ErrNotFound, errStr):
		return true, ErrExists
	case CorrespondTo(ErrUnsupported, errStr):
		return true, ErrUnsupported
	case CorrespondTo(ErrUnavailable, errStr):
		return true, ErrUnavailable
	case CorrespondTo(ErrWrongUser, errStr):
		return true, ErrWrongUser
	case CorrespondTo(ErrUnauthorised, errStr):
		return true, ErrUnauthorised
	case CorrespondTo(ErrUnknown, errStr):
		return true, ErrUnknown
	case CorrespondTo(ErrInvalid, errStr):
		return true, ErrInvalid
	case CorrespondTo(ErrConflict, errStr):
		return true, ErrConflict
	case CorrespondTo(ErrMarshalling, errStr):
		return true, ErrMarshalling
	case CorrespondTo(ErrCancelled, errStr):
		return true, ErrCancelled
	case CorrespondTo(ErrEmpty, errStr):
		return true, ErrEmpty
	case CorrespondTo(ErrUnexpected, errStr):
		return true, ErrUnexpected
	case CorrespondTo(ErrTooLarge, errStr):
		return true, ErrTooLarge
	case CorrespondTo(ErrForbidden, errStr):
		return true, ErrForbidden
	case CorrespondTo(ErrCondition, errStr):
		return true, ErrCondition
	case CorrespondTo(ErrEOF, errStr):
		return true, ErrEOF
	case CorrespondTo(ErrMalicious, errStr):
		return true, ErrMalicious
	case CorrespondTo(ErrWarning, errStr):
		return true, ErrWarning
	case CorrespondTo(ErrOutOfRange, errStr):
		return true, ErrOutOfRange
	}
	return false, ErrUnknown
}

// ConvertContextError converts a context error into common errors.
func ConvertContextError(err error) error {
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

// IsWarning will return whether an error is actually a warning
func IsWarning(target error) bool {
	if target == nil {
		return false
	}

	if Any(target, ErrWarning) {
		return true
	}

	underlyingErr := errors.Unwrap(target)
	if underlyingErr == nil {
		return false
	}

	return strings.TrimSuffix(target.Error(), underlyingErr.Error()) == warningStrPrepend
}

// NewWarning will create a warning wrapper around an existing commonerror so that it can be easily recovered. If the
// underlying error is not a commonerror then ok will be set to false
func NewWarning(target error) (ok bool, err error) {
	if target == nil {
		return false, nil
	}

	if !IsCommonError(target) {
		return false, target
	}

	if IsWarning(target) {
		return true, target
	}

	return true, fmt.Errorf("%v%w", warningStrPrepend, target)
}

// ParseWarning will extract the error that has been wrapped by ErrWarning. It will return nil if the error was not
// one of ErrWarning with ok set to false. It will also set ok to false if the underlying error cannot be parsed
func ParseWarning(target error) (ok bool, err error) {
	if target == nil || !IsWarning(target) {
		return
	}

	return true, errors.Unwrap(target)
}

// Ignore will return nil if the target error matches one of the errors to ignore
func Ignore(target error, ignore ...error) error {
	if Any(target, ignore...) {
		return nil
	}
	return target
}

// IgnoreCorrespondTo will return nil if the target error matches one of the errors to ignore
func IgnoreCorrespondTo(target error, ignore ...string) error {
	if CorrespondTo(target, ignore...) {
		return nil
	}
	return target
}

// IsEmpty states whether an error is empty or not.
// An error is considered empty if it is `nil` or has no description.
func IsEmpty(err error) bool {
	if err == nil {
		return true
	}
	if strings.TrimSpace(err.Error()) == "" {
		return true
	}
	return false
}

// Errorf is similar to fmt.Errorf although it will try to follow the error convention we use i.e. `errortype: message` but differs in that the wrapped error will be the targetErr
func Errorf(targetErr error, format string, args ...any) error {
	tErr := ConvertContextError(targetErr)
	if tErr == nil {
		tErr = ErrUnknown
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	cleansedMsg := strings.TrimSpace(msg)
	if cleansedMsg == "" {
		return tErr
	} else {
		return fmt.Errorf("%w%v %v", tErr, string(TypeReasonErrorSeparator), cleansedMsg)
	}
}

// WrapError wraps an error into a particular targetError. However, if the original error has to do with a contextual error (i.e. ErrCancelled or ErrTimeout), it will be passed through without having is type changed.
// This method should be used to safely wrap errors without losing information about context control information.
// If the target error is not set, the wrapped error will be of type ErrUnknown.
func WrapError(targetError, originalError error, msg string) error {
	tErr := targetError
	if tErr == nil {
		tErr = ErrUnknown
	}
	origErr := ConvertContextError(originalError)
	if Any(origErr, ErrTimeout, ErrCancelled) {
		tErr = origErr
	}
	if originalError == nil {
		return New(tErr, msg)
	} else {
		cleansedMsg := strings.TrimSpace(msg)
		if cleansedMsg == "" {
			return New(tErr, originalError.Error())
		} else {
			return Errorf(tErr, "%v%v %v", cleansedMsg, string(TypeReasonErrorSeparator), originalError.Error())
		}
	}
}

// WrapIfNotCommonError is similar to WrapError but only wraps an error if it is not a common error.
func WrapIfNotCommonError(targetError, originalError error, msg string) error {
	if Any(ConvertContextError(targetError), ErrTimeout, ErrCancelled) {
		return WrapError(targetError, originalError, msg)
	}
	if IsCommonError(originalError) {
		return New(originalError, msg)
	}
	return WrapError(targetError, originalError, msg)
}

// WrapErrorf is similar to WrapError but uses a format for the message
func WrapErrorf(targetError, originalError error, msgFormat string, args ...any) error {
	if len(args) == 0 {
		return WrapError(targetError, originalError, msgFormat)
	}
	return WrapError(targetError, originalError, fmt.Sprintf(msgFormat, args...))
}

// WrapIfNotCommonErrorf is similar to WrapError but only wraps an error if it is not a common error.
func WrapIfNotCommonErrorf(targetError, originalError error, msgFormat string, args ...any) error {
	if Any(ConvertContextError(targetError), ErrTimeout, ErrCancelled) {
		return WrapErrorf(targetError, originalError, msgFormat, args...)
	}
	if IsCommonError(originalError) {
		return Newf(originalError, msgFormat, args...)
	}
	return WrapErrorf(targetError, originalError, msgFormat, args...)
}

// New is similar to errors.New or fmt.Errorf but creates an error of type targetError
func New(targetError error, msg string) error {
	return Errorf(targetError, msg)
}

// Newf is similar to New but allows to format the message
func Newf(targetError error, msgFormat string, args ...any) error {
	return WrapErrorf(targetError, nil, msgFormat, args...)
}

// UndefinedVariable returns an undefined error related to a variable.
func UndefinedVariable(variableName string) error {
	return undefinedVariable(variableName, "")
}

// UndefinedVariableWithMessage returns an undefined error with a message.
func UndefinedVariableWithMessage(variableName string, msg string) error {
	return undefinedVariable(variableName, msg)
}

// UndefinedParameter returns an undefined error with a message
func UndefinedParameter(msg string) error {
	return undefinedVariable("", msg)
}

func undefinedVariable(variableName, msg string) error {
	if msg == "" {
		return Newf(ErrUndefined, "missing %v", variableName)
	}
	if variableName == "" {
		return New(ErrUndefined, msg)
	}
	return Newf(ErrUndefined, "missing %v: %v", variableName, msg)
}
