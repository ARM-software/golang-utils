/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package commonerrors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestAny(t *testing.T) {
	assert.True(t, Any(ErrNotImplemented, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.False(t, Any(ErrNotImplemented, ErrInvalid, ErrUnknown))
	assert.True(t, Any(ErrNotImplemented, nil, ErrNotImplemented))
	assert.True(t, Any(nil, nil, ErrNotImplemented))
	assert.False(t, Any(ErrNotImplemented, nil, ErrInvalid, ErrUnknown))
	assert.False(t, Any(nil, ErrInvalid, ErrUnknown))
	assert.True(t, Any(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.False(t, Any(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrUnknown))
}

func TestNone(t *testing.T) {
	assert.False(t, None(ErrNotImplemented, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.False(t, None(ErrNotImplemented, nil, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.True(t, None(ErrNotImplemented, ErrInvalid, ErrUnknown))
	assert.True(t, None(ErrNotImplemented, nil, ErrInvalid, ErrUnknown))
	assert.True(t, None(nil, ErrInvalid, ErrUnknown))
	assert.False(t, None(nil, nil, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.False(t, None(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.True(t, None(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrUnknown))
}

func TestCorrespondTo(t *testing.T) {
	assert.False(t, CorrespondTo(nil))
	assert.False(t, CorrespondTo(nil, faker.Sentence()))
	assert.False(t, CorrespondTo(ErrNotImplemented, ErrInvalid.Error(), ErrUnknown.Error()))
	assert.True(t, CorrespondTo(ErrNotImplemented, ErrInvalid.Error(), ErrNotImplemented.Error()))
	assert.True(t, CorrespondTo(fmt.Errorf("%v %w", faker.Sentence(), ErrUndefined), ErrUndefined.Error()))
	assert.True(t, CorrespondTo(fmt.Errorf("%v %v", faker.Sentence(), strings.ToUpper(ErrUndefined.Error())), strings.ToLower(ErrUndefined.Error())))
}

func TestIgnoreCorrespondTo(t *testing.T) {
	assert.NoError(t, IgnoreCorrespondTo(errors.New("test"), "test"))
	assert.NoError(t, IgnoreCorrespondTo(errors.New("test 123"), "test"))
	assert.Error(t, IgnoreCorrespondTo(errors.New("test 123"), "abc", "def", faker.Word()))
	assert.NoError(t, IgnoreCorrespondTo(ErrCondition, "condition"))
}

func TestContextErrorConversion(t *testing.T) {
	defer goleak.VerifyNone(t)
	task := func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Asked to stop:", ctx.Err())
				return
			default:
				time.Sleep(time.Second * 1)
			}
		}
	}
	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	go task(cancelCtx)
	time.Sleep(time.Second * 3)
	cancelFunc()
	time.Sleep(time.Second * 1)
	err := ConvertContextError(cancelCtx.Err())
	require.NotNil(t, err)
	assert.True(t, Any(err, ErrTimeout, ErrCancelled))
}

func TestIsCommonError(t *testing.T) {
	commonErrors := []error{
		ErrNotImplemented,
		ErrNoExtension,
		ErrNoLogger,
		ErrNoLoggerSource,
		ErrNoLogSource,
		ErrUndefined,
		ErrInvalidDestination,
		ErrTimeout,
		ErrLocked,
		ErrStaleLock,
		ErrExists,
		ErrNotFound,
		ErrUnsupported,
		ErrUnavailable,
		ErrWrongUser,
		ErrUnauthorised,
		ErrUnknown,
		ErrInvalid,
		ErrConflict,
		ErrMarshalling,
		ErrCancelled,
		ErrEmpty,
		ErrUnexpected,
		ErrTooLarge,
		ErrForbidden,
		ErrCondition,
		ErrEOF,
		ErrMalicious,
		ErrWarning,
		ErrOutOfRange,
		ErrFailed,
	}
	for i := range commonErrors {
		assert.True(t, IsCommonError(commonErrors[i]))
	}

	assert.False(t, IsCommonError(errors.New(faker.Sentence())))
}

func TestIsWarning(t *testing.T) {
	assert.True(t, IsWarning(ErrWarning))
	assert.False(t, IsWarning(ErrUnexpected))
	assert.False(t, IsWarning(nil))
	assert.True(t, IsWarning(fmt.Errorf("%w: i am i warning", ErrWarning)))
	assert.True(t, IsWarning(fmt.Errorf("%w: i am i warning too: %v", ErrWarning, ErrUnknown)))
}

func TestNewWarning(t *testing.T) {
	testErr := fmt.Errorf("%w: i am a test error", ErrUnexpected)

	t.Run("Normal", func(t *testing.T) {
		ok, err := NewWarning(testErr)
		assert.True(t, ok)
		assert.Equal(t, fmt.Errorf("%v%w", warningStrPrepend, testErr), err)
	})

	t.Run("Nil", func(t *testing.T) {
		ok, err := NewWarning(nil)
		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("Not commonerror", func(t *testing.T) {
		fakeError := errors.New(faker.Word())
		ok, err := NewWarning(fakeError)
		assert.False(t, ok)
		assert.Equal(t, fakeError, err)
	})

	t.Run("Warning on a warning", func(t *testing.T) {
		ok, err := NewWarning(testErr)
		assert.True(t, ok)
		ok, err = NewWarning(err)
		assert.True(t, ok)
		assert.Equal(t, fmt.Errorf("%v%w", warningStrPrepend, testErr), err)
	})
}

func TestParseWarning(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		testErr := fmt.Errorf("%w: i am a test error", ErrUnexpected)
		ok, errWarning := NewWarning(testErr)
		require.True(t, ok)
		require.True(t, IsWarning(errWarning))
		ok, err := ParseWarning(errWarning)
		assert.True(t, ok)
		assert.Equal(t, testErr, err)
	})

	t.Run("Nil", func(t *testing.T) {
		ok, err := ParseWarning(nil)
		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("Not Warning", func(t *testing.T) {
		testErr := fmt.Errorf("%w: i am a test error", ErrUnexpected)
		ok, err := ParseWarning(testErr)
		assert.False(t, ok)
		assert.Nil(t, err)
	})
}

func TestIgnore(t *testing.T) {
	assert.Equal(t, nil, Ignore(ErrNotImplemented, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.Equal(t, ErrNotImplemented, Ignore(ErrNotImplemented, ErrInvalid, ErrUnknown))
	assert.Equal(t, nil, Ignore(ErrNotImplemented, nil, ErrNotImplemented))
	assert.Equal(t, nil, Ignore(nil, nil, ErrNotImplemented))
	assert.Equal(t, ErrNotImplemented, Ignore(ErrNotImplemented, nil, ErrInvalid, ErrUnknown))
	assert.Equal(t, nil, Ignore(nil, ErrInvalid, ErrUnknown))
	assert.Equal(t, nil, Ignore(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.Equal(t, fmt.Errorf("an error %w", ErrNotImplemented), Ignore(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrUnknown))
}

func TestErrorRelatesTo(t *testing.T) {
	assert.True(t, RelatesTo(fmt.Errorf("%w: %v", ErrInvalid, faker.Sentence()).Error(), ErrInvalid))
	assert.True(t, RelatesTo(fmt.Errorf("%w: %v", ErrInvalid, faker.Sentence()).Error(), ErrInvalid, ErrCondition))
	assert.False(t, RelatesTo(fmt.Errorf("%w: %v", ErrUnauthorised, faker.Sentence()).Error(), ErrInvalid))
	assert.False(t, RelatesTo(fmt.Errorf("%w: %v", ErrUnauthorised, faker.Sentence()).Error(), ErrInvalid, ErrCondition))
	assert.True(t, RelatesTo(fmt.Sprint(fmt.Errorf("%w: %v", ErrInvalid, faker.Sentence())), ErrInvalid))
	assert.True(t, RelatesTo(fmt.Sprint(fmt.Errorf("%w: %v", ErrInvalid, faker.Sentence())), ErrInvalid, ErrCondition))
	assert.False(t, RelatesTo(fmt.Sprint(fmt.Errorf("%w: %v", ErrUnauthorised, faker.Sentence())), ErrInvalid))
	assert.False(t, RelatesTo(fmt.Sprint(fmt.Errorf("%w: %v", ErrUnauthorised, faker.Sentence())), ErrInvalid, ErrCondition))
	assert.True(t, RelatesTo(fmt.Sprintf("%v: %v", ErrInvalid.Error(), faker.Sentence()), ErrInvalid))
	assert.True(t, RelatesTo(fmt.Sprintf("%v: %v", ErrInvalid.Error(), faker.Sentence()), ErrInvalid, ErrCondition))
	assert.False(t, RelatesTo(fmt.Sprintf("%v: %v", ErrUnauthorised.Error(), faker.Sentence()), ErrInvalid))
	assert.False(t, RelatesTo(fmt.Sprintf("%v: %v", ErrUnauthorised.Error(), faker.Sentence()), ErrInvalid, ErrCondition))
}

func TestWrapError(t *testing.T) {
	assert.True(t, Any(WrapError(ErrUndefined, nil, faker.Sentence()), ErrUndefined))
	assert.True(t, Any(WrapError(ErrUndefined, ErrNotFound, faker.Sentence()), ErrUndefined))
	assert.True(t, Any(WrapError(nil, ErrNotFound, faker.Sentence()), ErrUnknown))
	assert.True(t, Any(WrapError(ErrUndefined, context.DeadlineExceeded, faker.Sentence()), ErrTimeout))
	assert.True(t, Any(WrapError(ErrUndefined, context.Canceled, faker.Sentence()), ErrCancelled))
	assert.True(t, Any(WrapError(ErrUndefined, ErrTimeout, faker.Sentence()), ErrTimeout))
	assert.True(t, Any(WrapError(ErrUndefined, ErrCancelled, faker.Sentence()), ErrCancelled))
	assert.True(t, Any(WrapErrorf(context.DeadlineExceeded, nil, faker.Sentence()), ErrTimeout))
	assert.True(t, Any(WrapError(context.Canceled, ErrConflict, faker.Sentence()), ErrCancelled))
	assert.True(t, Any(WrapErrorf(context.DeadlineExceeded, nil, "%v this is a test %v", faker.Name(), faker.Word()), ErrTimeout))
	assert.True(t, Any(Newf(context.DeadlineExceeded, "%v this is a test %v", faker.Name(), faker.Word()), ErrTimeout))
	assert.True(t, Any(WrapIfNotCommonError(context.Canceled, ErrConflict, faker.Sentence()), ErrCancelled))
	assert.True(t, Any(WrapIfNotCommonError(ErrUndefined, ErrConflict, faker.Sentence()), ErrConflict))
	assert.True(t, Any(WrapIfNotCommonErrorf(ErrUndefined, ErrConflict, faker.Sentence()), ErrConflict))
	assert.True(t, Any(WrapIfNotCommonError(ErrUndefined, errors.New(faker.Sentence()), faker.Sentence()), ErrUndefined))
	assert.True(t, Any(WrapIfNotCommonErrorf(ErrUndefined, errors.New(faker.Sentence()), faker.Sentence()), ErrUndefined))
}

func TestString(t *testing.T) {
	assert.Equal(t, "unknown", New(nil, "").Error())
	assert.Equal(t, "unknown", Newf(nil, "").Error())
	assert.Equal(t, "unknown", WrapError(nil, nil, "").Error())
	assert.Equal(t, "unknown", WrapErrorf(nil, nil, "").Error())
	assert.Equal(t, "unsupported", New(ErrUnsupported, "").Error())
	assert.Equal(t, "unsupported", Newf(ErrUnsupported, "").Error())
	assert.Equal(t, "unsupported", WrapError(ErrUnsupported, nil, "").Error())
	assert.Equal(t, "unsupported", WrapErrorf(ErrUnsupported, nil, "").Error())
	assert.Equal(t, "unsupported: test", New(ErrUnsupported, "test").Error())
	assert.Equal(t, "unknown: test", New(nil, "test").Error())
	assert.Equal(t, "unsupported: test 56", Newf(ErrUnsupported, "test %v", 56).Error())
	assert.Equal(t, "unsupported: test", WrapError(ErrUnsupported, nil, "test").Error())
	assert.Equal(t, "unsupported: not found", WrapError(ErrUnsupported, ErrNotFound, "").Error())
	assert.Equal(t, "unknown: test: unsupported", WrapError(nil, ErrUnsupported, "test").Error())
	assert.Equal(t, "unsupported: test 56: not found", WrapErrorf(ErrUnsupported, ErrNotFound, "test %v", 56).Error())
}
