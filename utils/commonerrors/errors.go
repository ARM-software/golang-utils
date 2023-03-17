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
