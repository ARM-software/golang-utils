package tus

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/hashing"
)

func TestParseTUSHash(t *testing.T) {
	sha1, err := hashing.DetermineHashingAlgorithm("sha1")
	require.NoError(t, err)

	tests := []struct {
		header           string
		expectedAlgo     string
		expectedChecksum string
		expectedError    error
	}{
		{
			header:           "sha1 MmFhZTZjMzVjOTRmY2ZiNDE1ZGJlOTVmNDA4YjljZTkxZWU4NDZlZA==",
			expectedAlgo:     hashing.HashSha1,
			expectedChecksum: hashing.CalculateStringHash(sha1, "hello world"),
		},
		{
			header:           "sha1 dGhpcyBpcyBhIHRlc3QgdmFsdWUgb2J2aW91c2x5==",
			expectedAlgo:     hashing.HashSha1,
			expectedChecksum: "this is a test value obviously",
		},
		{
			header:           "sha256 dGhpcyBpcyBhIHRlc3QgdmFsdWUgb2J2aW91c2x5=",
			expectedAlgo:     hashing.HashSha256,
			expectedChecksum: "this is a test value obviously",
		},
		{
			header:           "sha1",
			expectedAlgo:     "",
			expectedChecksum: "",
			expectedError:    commonerrors.ErrInvalid,
		},
		{
			header:           " Lve95gjOVATpfV8EL5X4nxwjKHE=",
			expectedAlgo:     "",
			expectedChecksum: "",
			expectedError:    commonerrors.ErrInvalid,
		},
		{
			header:           "",
			expectedAlgo:     "",
			expectedChecksum: "",
			expectedError:    commonerrors.ErrUndefined,
		},
		{
			header:           "sha1   dGhpcyBpcyBhIHRlc3QgdmFsdWUgb2J2aW91c2x5",
			expectedAlgo:     hashing.HashSha1,
			expectedChecksum: "this is a test value obviously",
		},
		{
			header:           "sha1 not_base64!!!",
			expectedAlgo:     "",
			expectedChecksum: "",
			expectedError:    commonerrors.ErrInvalid,
		},
		{
			header:           "SHA256 dGhpcyBpcyBhIHRlc3QgdmFsdWUgb2J2aW91c2x5=",
			expectedAlgo:     hashing.HashSha256,
			expectedChecksum: "this is a test value obviously",
		},
		{
			header:           "sha1-md5 Lve95gjOVATpfV8EL5X4nxwjKHE=",
			expectedAlgo:     "sha1-md5",
			expectedChecksum: "Lve95gjOVATpfV8EL5X4nxwjKHE=",
			expectedError:    commonerrors.ErrUnsupported,
		},
		{
			header:           "sha1 Lve95g jOVATpfV8EL5X4nxwjKHE=",
			expectedAlgo:     "",
			expectedChecksum: "",
			expectedError:    commonerrors.ErrInvalid,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.header, func(t *testing.T) {
			algo, hash, err := ParseTUSHash(test.header)

			if test.expectedError != nil {
				errortest.AssertError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedAlgo, algo)
				assert.Equal(t, test.expectedChecksum, hash)
			}
		})
	}
}

func TestParseTUSConcatHeader(t *testing.T) {
	url1 := strings.ToLower(faker.URL())
	url2 := strings.ToLower(faker.URL())

	tests := []struct {
		input              string
		isPartial          bool
		expectedPartialURL []string
		expectedError      error
	}{
		{
			input:              "partial",
			isPartial:          true,
			expectedPartialURL: nil,
		},
		{
			input:     "final; https://example.com/uploads/1 /files/2",
			isPartial: false,
			expectedPartialURL: []string{
				"https://example.com/uploads/1",
				"/files/2",
			},
		},
		{
			input:     "final; /a   /b",
			isPartial: false,
			expectedPartialURL: []string{
				"/a",
				"/b",
			},
		},
		{
			input:     "   final;   /x /y   ",
			isPartial: false,
			expectedPartialURL: []string{
				"/x",
				"/y",
			},
		},
		{
			input:     fmt.Sprintf("   final;   %v %v   ", url1, url2),
			isPartial: false,
			expectedPartialURL: []string{
				url1,
				url2,
			},
		},
		{
			input:     "final; /u?id=1#frag /v?w=2",
			isPartial: false,
			expectedPartialURL: []string{
				"/u?id=1#frag",
				"/v?w=2",
			},
		},
		{
			input:         "final;",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "final",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "",
			expectedError: commonerrors.ErrUndefined,
		},
		{
			input:         "partial; /a /b",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         fmt.Sprintf("%v; /a /b", faker.Word()),
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "final; /good /bad %url",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "final; /a\t/b",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "final; http://%zz",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:     "final;  /a    ",
			isPartial: false,
			expectedPartialURL: []string{
				"/a",
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.input, func(t *testing.T) {
			isPartial, partials, err := ParseTUSConcatHeader(test.input)

			if test.expectedError != nil {
				errortest.AssertError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.isPartial, isPartial)
				p := collection.Map[*url.URL, string](partials, func(u *url.URL) string {
					require.NotNil(t, u)
					return strings.ToLower(u.String())
				})
				assert.ElementsMatch(t, test.expectedPartialURL, p)
			}
		})
	}

}

func toBase64Encoded(s string) string { return base64.URLEncoding.EncodeToString([]byte(s)) }

func TestParseTUSMetadataHeader(t *testing.T) {

	key := faker.Word()
	value := faker.Paragraph()
	utf8Name := "résumé 2025.txt"
	valuea := faker.Paragraph()
	valueb := faker.Paragraph()

	tests := []struct {
		input            string
		expectedFilename *string
		expectedElements map[string]any
		expectedError    error
	}{
		{
			input:         "         ",
			expectedError: commonerrors.ErrUndefined,
		},
		{
			input:            `  filename d29ybGRfZG9taW5hdGlvbl9wbGFuLnBkZg==,is_confidential`,
			expectedFilename: field.ToOptionalString("world_domination_plan.pdf"),
			expectedElements: map[string]any{
				"filename":        "world_domination_plan.pdf",
				"is_confidential": true,
			},
		},
		{
			input: fmt.Sprintf("%v %v", key, toBase64Encoded(value)),
			expectedElements: map[string]any{
				key: value,
			},
		},
		{
			input:            fmt.Sprintf("foo,bar,filename %v", toBase64Encoded("x")),
			expectedFilename: field.ToOptionalString("x"),
			expectedElements: map[string]any{
				"foo":      true, // empty value
				"bar":      true, // empty value
				"filename": "x",
			},
		},
		{
			input:            fmt.Sprintf("filename %v , meta %v, empty", toBase64Encoded("x"), toBase64Encoded("y")),
			expectedFilename: field.ToOptionalString("x"),
			expectedElements: map[string]any{
				"filename": "x",
				"meta":     "y",
				"empty":    true,
			},
		},
		{
			input: "note " + toBase64Encoded("A/B+C=D=="),
			expectedElements: map[string]any{
				"note": "A/B+C=D==",
			},
		},
		{
			input:            "filename " + toBase64Encoded(utf8Name),
			expectedFilename: field.ToOptionalString(utf8Name),
			expectedElements: map[string]any{
				"filename": utf8Name,
			},
		},
		{
			input: fmt.Sprintf("      %v  ", toBase64Encoded("x")),
			expectedElements: map[string]any{
				toBase64Encoded("x"): true,
			},
		},
		{
			input:         "file name " + toBase64Encoded("x"),
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "a " + toBase64Encoded("1") + ",a " + toBase64Encoded("2"),
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input:         "filename not-base64@@",
			expectedError: commonerrors.ErrInvalid,
		},
		{
			input: fmt.Sprintf("a %v,,b %v", toBase64Encoded(valuea), toBase64Encoded(valueb)),
			expectedElements: map[string]any{
				"a": valuea,
				"b": valueb,
			},
		},
		{
			input: fmt.Sprintf("a %v,b %v,", toBase64Encoded(valuea), toBase64Encoded(valueb)),
			expectedElements: map[string]any{
				"a": valuea,
				"b": valueb,
			},
		},
		{
			input:            "   , filename " + toBase64Encoded("x"),
			expectedFilename: field.ToOptionalString("x"),
			expectedElements: map[string]any{
				"filename": "x",
			},
		},
		{
			input: func() string {
				parts := []string{
					"id " + toBase64Encoded("123"),
					"owner " + toBase64Encoded("alice"),
					"tag " + toBase64Encoded("blue"),
					"desc " + toBase64Encoded("hello world"),
					"filename " + toBase64Encoded("big.bin"),
				}
				return strings.Join(parts, ", ")
			}(),
			expectedFilename: field.ToOptionalString("big.bin"),
			expectedElements: map[string]any{
				"id":       "123",
				"owner":    "alice",
				"tag":      "blue",
				"desc":     "hello world",
				"filename": "big.bin",
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.input, func(t *testing.T) {
			filename, elems, err := ParseTUSMetadataHeader(test.input)

			if test.expectedError != nil {
				errortest.AssertError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, field.OptionalString(test.expectedFilename, ""), field.OptionalString(filename, ""))
				actualElements := collection.ConvertMapToPairSlice(elems, "=")
				expectedElements := collection.ConvertMapToPairSlice(test.expectedElements, "=")
				assert.ElementsMatch(t, expectedElements, actualElements)
			}
		})
	}
}
