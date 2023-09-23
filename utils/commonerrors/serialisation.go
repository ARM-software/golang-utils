package commonerrors

import (
	"errors"
	"fmt"
	"strings"
)

type marshallingError struct {
	Reason string
	Error  error
}

func (e *marshallingError) MarshalText() (text []byte, err error) {
	str := serialiseMarshallingError(e)
	return []byte(str), nil
}

func (e *marshallingError) UnmarshalText(text []byte) error {
	er := processErrorStr(string(text))
	if er == nil {
		return ErrMarshalling
	}
	e.Error = er.Error
	e.Reason = er.Reason
	return nil
}

func (e *marshallingError) ConvertToError() error {
	if e == nil {
		return nil
	}
	if e.Error == nil {
		if e.Reason == "" {
			return nil
		}
		return errors.New(e.Reason)
	}
	if e.Reason == "" {
		return e.Error
	}
	return fmt.Errorf("%w: %v", e.Error, e.Reason)
}

func processError(err error) (mErr *marshallingError) {
	if err == nil {
		return
	}
	mErr = processErrorStr(err.Error())
	if mErr == nil {
		mErr = &marshallingError{
			Error: err,
		}
		return
	}
	switch x := err.(type) {
	case interface{ Unwrap() error }:
		mErr.Error = x.Unwrap()
	case interface{ Unwrap() []error }:
		mErr.Error = errors.Join(x.Unwrap()...)
	}
	return
}

func processErrorStr(err string) (mErr *marshallingError) {
	err = strings.TrimSpace(err)
	if err == "" {
		return
	}
	mErr = &marshallingError{}
	elems := strings.Split(err, ":")
	found, commonErr := deserialiseCommonError(elems[0])
	if !found || commonErr == nil {
		mErr.Error = errors.New(strings.TrimSpace(elems[0]))
	} else {
		mErr.Error = commonErr
	}
	if len(elems) > 0 {
		var reasonElems []string
		for i := 1; i < len(elems); i++ {
			reasonElems = append(reasonElems, strings.TrimSpace(elems[i]))
		}
		mErr.Reason = strings.Join(reasonElems, ": ")
	}
	return
}

func serialiseMarshallingError(err *marshallingError) string {
	if err == nil {
		return ""
	}
	mErr := err.ConvertToError()
	if mErr == nil {
		return ""
	}
	return mErr.Error()
}

// SerialiseError marshals an error following a certain convention: `error type: reason`
func SerialiseError(err error) ([]byte, error) {
	mErr := processError(err)
	if mErr == nil {
		return nil, nil
	}
	return mErr.MarshalText()
}

// DeserialiseError unmarshals text into an error. It tries to determine the error type.
func DeserialiseError(text []byte) (deserialisedError, err error) {
	if len(text) == 0 {
		return
	}
	mErr := marshallingError{}
	err = mErr.UnmarshalText(text)
	if err != nil {
		return
	}
	deserialisedError = mErr.ConvertToError()
	return
}
