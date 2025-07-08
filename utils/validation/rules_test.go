package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestCastingToInt(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		err   error
	}{
		{"int", int(8080), nil},
		{"int8", int8(80), nil},
		{"int16", int16(8080), nil},
		{"int32", int32(8080), nil},
		{"int64", int64(8080), nil},
		{"uint", uint(8080), nil},
		{"uint8", uint8(80), nil},
		{"uint16", uint16(8080), nil},
		{"uint32", uint32(8080), nil},
		{"uint64", uint64(8080), nil},
		{"string valid", "8080", nil},
		{"[]byte valid", []byte("8080"), nil},
		{"int min valid port", int(1), nil},
		{"int max valid port", int(65535), nil},
		{"string min valid port", "1", nil},
		{"string max valid port", "65535", nil},
		{"int below range", int(0), commonerrors.ErrInvalid},
		{"int above range", int(65536), commonerrors.ErrInvalid},
		{"uint above range", uint(65536), commonerrors.ErrInvalid},
		{"string negative", "-1", commonerrors.ErrInvalid},
		{"string above range", "65536", commonerrors.ErrInvalid},
		{"string non-numeric", "notaport", commonerrors.ErrInvalid},
		{"[]byte non-numeric", []byte("notaport"), commonerrors.ErrInvalid},
		{"float64", float64(8080), commonerrors.ErrMarshalling},
		{"bool", true, commonerrors.ErrMarshalling},
		{"struct", struct{}{}, commonerrors.ErrMarshalling},
		{"nil", nil, commonerrors.ErrMarshalling},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := IsPort.Validate(test.value)
			if test.err == nil {
				assert.NoError(t, err)
			} else {
				errortest.AssertError(t, err, test.err)
			}
		})
	}
}
