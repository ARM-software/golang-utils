package yaml

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestIsValidYAML(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		err   error
	}{
		{name: "valid string", value: "name: value\ncount: 2\n"},
		{name: "valid bytes", value: []byte("- name: value\n")},
		{name: "invalid string", value: "name: [value\n", err: commonerrors.ErrInvalid},
		{name: "invalid bytes", value: []byte("name: [value\n"), err: commonerrors.ErrInvalid},
		{name: "unsupported type", value: 123, err: commonerrors.ErrMarshalling},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validation.Validate(test.value, IsValidYAML)
			if test.err == nil {
				assert.NoError(t, err)
				return
			}
			errortest.AssertError(t, err, test.err)
		})
	}
}
