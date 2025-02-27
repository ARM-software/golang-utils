package commonerrors

import (
	"encoding"
	"errors"
	"fmt"
	"strings"
)

const (
	TypeReasonErrorSeparator = ':'
	MultipleErrorSeparator   = '\n'
)

type iMarshallingError interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	fmt.Stringer
	error
	ConvertToError() error
	SetWrappedError(err error)
}

type marshallingError struct {
	Reason    string
	ErrorType error
}

func (e *marshallingError) MarshalText() (text []byte, err error) {
	str := e.String()
	return []byte(str), nil
}

func (e *marshallingError) String() string {
	return serialiseMarshallingError(e)
}

func (e *marshallingError) Error() string {
	return e.String()
}

func (e *marshallingError) UnmarshalText(text []byte) error {
	er := processErrorStrLine(string(text))
	if er == nil {
		return ErrMarshalling
	}
	e.ErrorType = er.ErrorType
	e.Reason = er.Reason
	return nil
}

func (e *marshallingError) ConvertToError() error {
	if e == nil {
		return nil
	}
	if e.ErrorType == nil {
		if e.Reason == "" {
			return nil
		}
		return errors.New(e.Reason)
	}
	if e.Reason == "" {
		return e.ErrorType
	}
	return New(e.ErrorType, e.Reason)
}

func (e *marshallingError) SetWrappedError(err error) {
	e.ErrorType = err
}

type multiplemarshallingError struct {
	subErrs []iMarshallingError
}

func (m *multiplemarshallingError) MarshalText() (text []byte, err error) {
	for i := range m.subErrs {
		subtext, suberr := m.subErrs[i].MarshalText()
		if suberr != nil {
			err = WrapError(ErrMarshalling, suberr, "an error item could not be marshalled")
			return
		}
		text = append(text, subtext...)
		text = append(text, MultipleErrorSeparator)
	}
	return
}

func (m *multiplemarshallingError) String() string {
	text, err := m.MarshalText()
	if err == nil {
		return string(text)
	}
	return ""
}

func (m *multiplemarshallingError) Error() string {
	return m.String()
}

func (m *multiplemarshallingError) UnmarshalText(text []byte) error {
	sub := processErrorStr(string(text))
	if IsEmpty(sub) {
		return ErrMarshalling
	}
	if mul, ok := sub.(*multiplemarshallingError); ok {
		m.subErrs = mul.subErrs
	} else {
		m.subErrs = append(m.subErrs, sub)
	}
	return nil
}

func (m *multiplemarshallingError) ConvertToError() error {
	var errs []error
	for i := range m.subErrs {
		errs = append(errs, m.subErrs[i].ConvertToError())
	}
	return errors.Join(errs...)
}

func (m *multiplemarshallingError) SetWrappedError(err error) {
	if err == nil {
		return
	}
	if x, ok := err.(interface{ Unwrap() []error }); ok {
		unwrapped := x.Unwrap()
		if len(unwrapped) > len(m.subErrs) {
			for i := 0; i < len(unwrapped)-len(m.subErrs); i++ {
				m.subErrs = append(m.subErrs, &marshallingError{})
			}
		}
		for i := range unwrapped {
			subErr := m.subErrs[i]
			if subErr != nil {
				subErr.SetWrappedError(unwrapped[i])
			}
		}
	}
}
func processErrorStr(s string) iMarshallingError {
	if strings.Contains(s, string(MultipleErrorSeparator)) {
		elems := strings.Split(s, string(MultipleErrorSeparator))
		m := &multiplemarshallingError{}
		for i := range elems {
			mErr := processErrorStrLine(elems[i])

			if mErr != nil {
				m.subErrs = append(m.subErrs, mErr)
			}
		}
		return m
	} else {
		return processErrorStrLine(s)
	}
}

func processError(err error) (mErr iMarshallingError) {
	if err == nil {
		return
	}
	mErr = processErrorStr(err.Error())
	if IsEmpty(mErr) {
		mErr = &marshallingError{
			ErrorType: Newf(ErrUnknown, "error `%T` with no description returned", err),
		}
		return
	}
	switch x := err.(type) {
	case interface{ Unwrap() error }:
		mErr.SetWrappedError(x.Unwrap())
	case interface{ Unwrap() []error }:
		unwrap := x.Unwrap()
		var nonNilUnwrappedErrors []error
		for i := range unwrap {
			if !IsEmpty(unwrap[i]) {
				nonNilUnwrappedErrors = append(nonNilUnwrappedErrors, unwrap[i])
			}
		}
		mErr.SetWrappedError(errors.Join(nonNilUnwrappedErrors...))
	}
	return
}

func processErrorStrLine(err string) (mErr *marshallingError) {
	err = strings.TrimSpace(err)
	if err == "" {
		return nil
	}
	mErr = &marshallingError{}
	elems := strings.Split(err, string(TypeReasonErrorSeparator))
	found, commonErr := deserialiseCommonError(elems[0])
	if !found || commonErr == nil {
		mErr.SetWrappedError(errors.New(strings.TrimSpace(elems[0])))
	} else {
		mErr.SetWrappedError(commonErr)
	}
	if len(elems) > 0 {
		var reasonElems []string
		for i := 1; i < len(elems); i++ {
			reasonElems = append(reasonElems, strings.TrimSpace(elems[i]))
		}
		mErr.Reason = strings.Join(reasonElems, fmt.Sprintf("%v ", string(TypeReasonErrorSeparator)))
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

// SerialiseError marshals an error following a certain convention: `error type: reason`.
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
	var mErr iMarshallingError

	if strings.Contains(string(text), string(MultipleErrorSeparator)) {
		mErr = &multiplemarshallingError{}
	} else {
		mErr = &marshallingError{}
	}
	err = mErr.UnmarshalText(text)
	if err != nil {
		return
	}
	deserialisedError = mErr.ConvertToError()
	return
}
