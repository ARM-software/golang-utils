package json

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestIsValidJSON(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		err   error
	}{
		{name: "valid string", value: `{"name":"value","count":2}`},
		{name: "valid bytes", value: []byte(`[{"name":"value"}]`)},
		{name: "invalid string", value: `{"name":`, err: commonerrors.ErrInvalid},
		{name: "invalid bytes", value: []byte(`{"name":`), err: commonerrors.ErrInvalid},
		{name: "unsupported type", value: true, err: commonerrors.ErrMarshalling},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validation.Validate(test.value, IsValidJSON)
			if test.err == nil {
				assert.NoError(t, err)
				return
			}
			errortest.AssertError(t, err, test.err)
		})
	}
}
