/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package commonerrors defines typical errors which can happen.
package commonerrors

import (
	"context"
	"errors"
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
)

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
//	  CorrespondTo(errors.New("feature a is not supported", "not supported") = True
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

// Ignore will return nil if the target error matches one of the errors to ignore
func Ignore(target error, ignore ...error) error {
	if Any(target, ignore...) {
		return nil
	}
	return target
}
