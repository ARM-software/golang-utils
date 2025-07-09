package base64

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// DecodeString decodes a base64 encoded string. An error is raised if decoding fails.
func DecodeString(ctx context.Context, s string) (decoded string, err error) {
	if reflection.IsEmpty(s) {
		err = commonerrors.New(commonerrors.ErrEmpty, "the string is empty")
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	data, err := base64.URLEncoding.DecodeString(s)
	if err == nil {
		decoded = string(data)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	data, err = base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		decoded = string(data)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	data, err = base64.StdEncoding.DecodeString(s)
	if err == nil {
		decoded = string(data)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	data, err = base64.RawStdEncoding.DecodeString(s)
	if err == nil {
		decoded = string(data)
	} else {
		trimmed := strings.TrimSuffix(strings.TrimSuffix(s, "="), "=")
		if trimmed == s || strings.HasSuffix(trimmed, "=") {
			err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to decode base64 string")
		} else {
			decoded, err = DecodeString(ctx, trimmed)
		}
	}
	return
}

// DecodeIfEncoded will attempt to decode any string if they are base64 encoded. If not, the string will be returned as is.
// If the string is base64 encoded but the decoding fails, the original string will be returned.
func DecodeIfEncoded(ctx context.Context, s string) (decoded string) {
	decoded = s
	if IsEncoded(s) {
		d, err := DecodeString(ctx, s)
		if err == nil {
			decoded = d
		}
	}
	return
}

// DecodeRecursively will attempt to decode any string until they are no longer base64 encoded.
func DecodeRecursively(ctx context.Context, s string) (decoded string) {
	decoded = s
	for {
		tmp := DecodeIfEncoded(ctx, decoded)
		if decoded == tmp {
			return
		}
		decoded = tmp
	}
}

// IsEncoded checks whether a string is encoded or not.
func IsEncoded(s string) bool {
	if reflection.IsEmpty(s) {
		return false
	}
	if is.Base64.Validate(s) == nil {
		return true
	}
	_, err := DecodeString(context.Background(), s)
	if err == nil {
		return true
	}
	trimmed := strings.TrimSuffix(strings.TrimSuffix(s, "="), "=")
	if trimmed == s || strings.HasSuffix(trimmed, "=") {
		return false
	} else {
		return IsEncoded(trimmed)
	}

}

func EncodeString(s string) string {
	return Encode([]byte(s))
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
