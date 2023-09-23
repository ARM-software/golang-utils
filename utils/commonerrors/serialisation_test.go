package commonerrors

import (
	"errors"
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestDeserialise(t *testing.T) {
	sentence := faker.Sentence()
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
			text:           fmt.Errorf("%w: %v", errors.New(errStr), sentence).Error(),
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
					assert.NoError(t, mErr.Error)
				} else {
					require.Error(t, mErr.Error)
					assert.Equal(t, test.expectedError.Error(), mErr.Error.Error())
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
			fmt.Println(dErr)
			assert.True(t, Any(dErr, test.commonError))
		})
	}
}

func TestGenericSerialisation(t *testing.T) {
	text, err := SerialiseError(errors.New(faker.Sentence()))
	require.NoError(t, err)
	assert.NotEmpty(t, text)
	deserialisedErr, err := DeserialiseError(text)
	require.NoError(t, err)
	assert.Error(t, deserialisedErr)
	assert.True(t, Any(deserialisedErr, ErrUnknown))

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
			reason := faker.Sentence()
			text, err := SerialiseError(fmt.Errorf("%w: %v", test.commonError, reason))
			require.NoError(t, err)
			dErr, err := DeserialiseError(text)
			require.NoError(t, err)
			assert.True(t, Any(dErr, test.commonError))
			fmt.Println(dErr)
			assert.True(t, strings.Contains(dErr.Error(), reason))
		})
	}
}
