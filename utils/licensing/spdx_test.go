package licensing

import (
	"net/url"
	"slices"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
)

func TestValidateSPDXLicence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		licence       string
		expectedError bool
	}{
		{"MIT", false},
		{"MIT OR Apache-2.0", false},
		{"GPL-2.0-or-later", false},
		{"", true},
		{"MIT OR", true},
		{"definitely-not-a-real-licence", true},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.licence, func(t *testing.T) {
			err := ValidateSPDXLicence(test.licence)

			if test.expectedError {
				errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsSPDXLicence(t *testing.T) {
	t.Parallel()

	t.Run("MIT", func(t *testing.T) {
		err := validation.Validate("MIT", IsSPDXLicence)
		require.NoError(t, err)
	})

	t.Run("MIT OR", func(t *testing.T) {
		err := validation.Validate("MIT OR", IsSPDXLicence)
		require.Error(t, err)
	})
}

func TestNormaliseSPDXLicence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expression     string
		expectedResult string
		expectedError  bool
	}{
		{"MIT", "MIT", false},
		{"MIT OR Apache-2.0", "MIT OR Apache-2.0", false},
		{"mit or apache 2", "MIT OR Apache-2.0", false},
		{"", "", true},
		{"MIT OR", "", true},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.expression, func(t *testing.T) {
			res, err := NormaliseSPDXLicence(test.expression)

			if test.expectedError {
				errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
				require.Empty(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedResult, res)
			}
		})
	}
}

func TestSatisfiesLicensingConstraints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		licence       string
		allowedList   []string
		expectedPass  bool
		expectedError bool
	}{
		{"MIT OR Apache-2.0", []string{"MIT"}, true, false},
		{"Apache-2.0", []string{"MIT"}, false, false},
		{"mit or apache 2", []string{"apache 2"}, true, false},
		{"MIT OR", []string{"MIT"}, false, true},
		{"MIT", []string{"Apache-2.0", "not-a-licence"}, false, true},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.licence, func(t *testing.T) {
			pass, err := SatisfiesLicensingConstraints(test.licence, test.allowedList)

			if test.expectedError {
				errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
				require.False(t, pass)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedPass, pass)
			}
		})
	}
}

func TestFetchLicenceURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		licence       string
		expectedURL   string
		expectedError bool
	}{
		{"MIT", "https://spdx.org/licenses/MIT.html", false},
		{"apache 2", "https://spdx.org/licenses/Apache-2.0.html", false},
		{"", "", true},
		{"", "", true},
		{"MIT OR Apache-2.0", "", true},
		{"not-a-licence", "", true},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.licence, func(t *testing.T) {
			u, err := FetchLicenceURL(field.ToOptionalStringOrNilIfEmpty(test.licence))
			if test.expectedError {
				errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
				require.Nil(t, u)
			} else {
				require.NoError(t, err)
				require.NotNil(t, u)
				require.Equal(t, test.expectedURL, u.String())
			}
		})
	}
}

func TestFetchLicenceURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expression    string
		expectedURLs  []string
		expectedError bool
	}{
		{
			"MIT",
			[]string{"https://spdx.org/licenses/MIT.html"},
			false,
		},
		{
			"MIT OR Apache-2.0",
			[]string{
				"https://spdx.org/licenses/MIT.html",
				"https://spdx.org/licenses/Apache-2.0.html",
			},
			false,
		},
		{
			"mit or apache 2",
			[]string{
				"https://spdx.org/licenses/MIT.html",
				"https://spdx.org/licenses/Apache-2.0.html",
			},
			false,
		},
		{"", nil, true},
		{"MIT OR", nil, true},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.expression, func(t *testing.T) {
			seq, err := FetchLicenceURLs(test.expression)

			if test.expectedError {
				errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
				require.Empty(t, seq)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, seq)
			require.ElementsMatch(t, test.expectedURLs, slices.Collect(collection.MapSequence[url.URL, string](seq, func(u url.URL) string { return u.String() })))
		})
	}
}
