package commonerrors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeserialise(t *testing.T) {
	sentence := strings.ReplaceAll(faker.Sentence(), "\n", ";")
	errStr := strings.ToLower(faker.Name())
	var tests = []struct {
		text             string
		expectedReason   string
		expectedError    error
		marshallingError bool
	}{
		{
			text:             "",
			expectedReason:   "",
			expectedError:    nil,
			marshallingError: true,
		},
		{
			text:           errStr,
			expectedReason: "",
			expectedError:  errors.New(errStr),
		},
		{
			text:           sentence,
			expectedReason: "",
			expectedError:  errors.New(sentence),
		},
		{
			text:           fmt.Errorf("%v:%v", errStr, sentence).Error(),
			expectedReason: sentence,
			expectedError:  errors.New(errStr),
		},
		{
			text:           New(errors.New(errStr), sentence).Error(),
			expectedReason: sentence,
			expectedError:  errors.New(errStr),
		},
		{
			text:           fmt.Errorf("%w : %v", errors.New(errStr), sentence).Error(),
			expectedReason: sentence,
			expectedError:  errors.New(errStr),
		},
		{
			text:           fmt.Errorf("%w : %v", ErrInvalid, sentence).Error(),
			expectedReason: sentence,
			expectedError:  ErrInvalid,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.text, func(t *testing.T) {
			mErr := marshallingError{}
			if test.marshallingError {
				assert.True(t, Any(ErrMarshalling, mErr.UnmarshalText([]byte(test.text))))
			} else {
				require.NoError(t, mErr.UnmarshalText([]byte(test.text)))
				if test.expectedError == nil {
					assert.NoError(t, mErr.ErrorType)
				} else {
					require.Error(t, mErr.ErrorType)
					assert.Equal(t, test.expectedError.Error(), mErr.ErrorType.Error())
				}

				assert.Equal(t, test.expectedReason, mErr.Reason)
			}
		})
	}
}

func TestCommonErrorSerialisation(t *testing.T) {
	text, err := SerialiseError(nil)
	require.NoError(t, err)
	assert.Empty(t, text)
	commonErr, err := DeserialiseError(nil)
	require.NoError(t, err)
	assert.NoError(t, commonErr)
	commonErr, err = DeserialiseError([]byte{})
	require.NoError(t, err)
	assert.NoError(t, commonErr)
	commonErr, err = DeserialiseError([]byte(""))
	require.NoError(t, err)
	assert.NoError(t, commonErr)
	tests := []struct {
		commonError error
	}{
		{commonError: ErrNotImplemented},
		{commonError: ErrNoExtension},
		{commonError: ErrNoLogger},
		{commonError: ErrNoLoggerSource},
		{commonError: ErrNoLogSource},
		{commonError: ErrUndefined},
		{commonError: ErrInvalidDestination},
		{commonError: ErrTimeout},
		{commonError: ErrLocked},
		{commonError: ErrStaleLock},
		{commonError: ErrExists},
		{commonError: ErrNotFound},
		{commonError: ErrUnsupported},
		{commonError: ErrUnavailable},
		{commonError: ErrWrongUser},
		{commonError: ErrUnauthorised},
		{commonError: ErrUnknown},
		{commonError: ErrInvalid},
		{commonError: ErrConflict},
		{commonError: ErrMarshalling},
		{commonError: ErrCancelled},
		{commonError: ErrEmpty},
		{commonError: ErrUnexpected},
		{commonError: ErrTooLarge},
		{commonError: ErrForbidden},
		{commonError: ErrCondition},
		{commonError: ErrEOF},
		{commonError: ErrMalicious},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.commonError.Error(), func(t *testing.T) {
			text, err := SerialiseError(test.commonError)
			require.NoError(t, err)
			found, dErr := deserialiseCommonError(string(text))
			assert.True(t, found)
			assert.True(t, Any(dErr, test.commonError))
			dErr, err = DeserialiseError(text)
			require.NoError(t, err)
			assert.True(t, Any(dErr, test.commonError))
		})
	}
}

type multiErr []error

func (m multiErr) Error() string   { return errors.Join(m...).Error() }
func (m multiErr) Unwrap() []error { return []error(m) }

func TestMultipleError(t *testing.T) {
	expectedErr := multiErr([]error{New(ErrInvalid, strings.ReplaceAll(faker.Sentence(), string(MultipleErrorSeparator), "sep")), errors.New(""), New(ErrUnexpected, strings.ReplaceAll(faker.Sentence(), string(MultipleErrorSeparator), ";"))})
	text, err := SerialiseError(expectedErr)
	require.NoError(t, err)
	assert.NotEmpty(t, text)

	deserialisedErr, err := DeserialiseError(text)
	require.NoError(t, err)
	assert.Error(t, deserialisedErr)
	subErrors := expectedErr.Unwrap()
	for i := range subErrors {
		assert.True(t, CorrespondTo(deserialisedErr, subErrors[i].Error()), subErrors[i].Error())
	}
}
func TestGenericSerialisation(t *testing.T) {
	t.Run("error with no type", func(t *testing.T) {
		expectedErr := errors.New(strings.ReplaceAll(faker.Sentence(), "\n", ";"))
		text, err := SerialiseError(expectedErr)
		require.NoError(t, err)
		assert.NotEmpty(t, text)
		deserialisedErr, err := DeserialiseError(text)
		require.NoError(t, err)
		assert.Error(t, deserialisedErr)
		assert.Equal(t, expectedErr.Error(), deserialisedErr.Error())
	})
	t.Run("no error", func(t *testing.T) {
		text, err := SerialiseError(nil)
		require.NoError(t, err)
		assert.Empty(t, text)
	})
	t.Run("error with no description", func(t *testing.T) {
		expectedErr := errors.New(" ")
		assert.True(t, IsEmpty(expectedErr))
		text, err := SerialiseError(expectedErr)
		require.NoError(t, err)
		assert.NotEmpty(t, text)
		deserialisedErr, err := DeserialiseError(text)
		require.NoError(t, err)
		assert.True(t, Any(deserialisedErr, ErrUnknown))
	})
	t.Run("error deserialisation", func(t *testing.T) {
		text := []byte("                      ")
		deserialisedErr, err := DeserialiseError(text)
		require.Error(t, err)
		assert.True(t, Any(err, ErrMarshalling))
		assert.NoError(t, deserialisedErr)
	})

	t.Run("error Reason", func(t *testing.T) {
		reason := faker.Sentence()
		dErr := New(errors.New(faker.Word()), reason)
		dReason, err := GetCommonErrorReason(dErr)
		require.NoError(t, err)
		assert.Equal(t, dErr.Error(), dReason)
		dReason, err = GetErrorReason(dErr)
		require.NoError(t, err)
		assert.Equal(t, reason, dReason)
		reason = faker.Sentence()
		dErr = New(ErrNotFound, reason)
		dReason, err = GetCommonErrorReason(dErr)
		require.NoError(t, err)
		assert.Equal(t, reason, dReason)
		dReason, err = GetErrorReason(dErr)
		require.NoError(t, err)
		assert.Equal(t, reason, dReason)
		dReason, err = GetCommonErrorReason(nil)
		assert.True(t, Any(err, ErrUndefined))
		assert.Empty(t, dReason)
		dReason, err = GetErrorReason(nil)
		assert.True(t, Any(err, ErrUndefined))
		assert.Empty(t, dReason)
	})

	tests := []struct {
		commonError error
	}{
		{commonError: ErrNotImplemented},
		{commonError: ErrNoExtension},
		{commonError: ErrNoLogger},
		{commonError: ErrNoLoggerSource},
		{commonError: ErrNoLogSource},
		{commonError: ErrUndefined},
		{commonError: ErrInvalidDestination},
		{commonError: ErrTimeout},
		{commonError: ErrLocked},
		{commonError: ErrStaleLock},
		{commonError: ErrExists},
		{commonError: ErrNotFound},
		{commonError: ErrUnsupported},
		{commonError: ErrUnavailable},
		{commonError: ErrWrongUser},
		{commonError: ErrUnauthorised},
		{commonError: ErrUnknown},
		{commonError: ErrInvalid},
		{commonError: ErrConflict},
		{commonError: ErrMarshalling},
		{commonError: ErrCancelled},
		{commonError: ErrEmpty},
		{commonError: ErrUnexpected},
		{commonError: ErrTooLarge},
		{commonError: ErrForbidden},
		{commonError: ErrCondition},
		{commonError: ErrEOF},
		{commonError: ErrMalicious},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.commonError.Error(), func(t *testing.T) {
			reason := strings.ReplaceAll(faker.Sentence(), "\n", ";")
			text, err := SerialiseError(New(test.commonError, reason))
			require.NoError(t, err)
			dErr, err := DeserialiseError(text)
			require.NoError(t, err)
			assert.True(t, Any(dErr, test.commonError))
			assert.True(t, strings.Contains(dErr.Error(), reason))
			underlyingErr, err := GetUnderlyingErrorType(dErr)
			require.NoError(t, err)
			assert.True(t, Any(underlyingErr, test.commonError))
			dReason, err := GetErrorReason(dErr)
			require.NoError(t, err)
			assert.Equal(t, reason, dReason)
			dReason, err = GetCommonErrorReason(dErr)
			require.NoError(t, err)
			assert.Equal(t, reason, dReason)
		})
	}
}
