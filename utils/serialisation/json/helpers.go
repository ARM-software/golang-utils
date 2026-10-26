// Package json mirrors the standard library's encoding/json helpers while
// adding two service-focused behaviours:
//   - stream helpers are context aware, so long reads and writes can be
//     cancelled through the supplied context
//   - marshal and unmarshal operations automatically use generated fast JSON
//     encoders and decoders when a type exposes those interfaces
//
// The intent is to keep call sites as close as possible to encoding/json while
// letting services benefit from non-reflective JSON code generation without
// each caller having to know which implementation a type uses. Examples of the
// supported fast-path libraries are github.com/mailru/easyjson and
// github.com/pquerna/ffjson.
package json

import (
	"bytes"
	"context"

	"github.com/mailru/easyjson"
	"github.com/pquerna/ffjson/ffjson"
	sigsyaml "sigs.k8s.io/yaml"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	nullBytes = []byte("null")
	// JSONExtensions is the list of file extensions that are considered JSON files.
	JSONExtensions = []string{".json"}
)

// Marshal encodes a value to a JSON byte slice.
// It matches encoding/json.Marshal.
//
// If the value implements a generated fast JSON marshaler interface, that
// non-reflective path is used. Otherwise, it falls back to the regular runtime
// path available through the supported fast JSON backend. The rationale is to
// keep one familiar helper for callers while automatically taking a faster
// serialisation path when generated code exists. Examples of supported
// generators are github.com/mailru/easyjson and github.com/pquerna/ffjson.
func Marshal(v any) ([]byte, error) {
	if reflection.IsNilInterface(v) {
		return nullBytes, nil
	}

	fastMarshaller, ok := v.(easyjson.Marshaler)
	if ok {
		return easyjson.Marshal(fastMarshaller)
	}

	return ffjson.Marshal(v)
}

// MarshalWithContext encodes a value to a JSON byte slice using a context-aware
// Encoder.
// It follows the same helper shape as Marshal, but routes the write through the
// package's context-aware streaming helpers.
func MarshalWithContext(ctx context.Context, v any) (content []byte, err error) {
	var buf bytes.Buffer
	encoder := NewEncoder(ctx, &buf)
	err = encoder.Encode(v)
	if err != nil {
		return
	}
	content = buf.Bytes()
	return
}

// Unmarshal decodes a JSON byte slice into a destination value.
// It matches encoding/json.Unmarshal.
//
// If the destination implements a generated fast JSON unmarshaler interface,
// that non-reflective path is used. Otherwise it falls back to the regular
// runtime path available through the supported fast JSON backend. This lets
// calling code stay implementation-agnostic while types that opt into
// generated JSON code get faster deserialisation automatically. Examples of
// supported generators are github.com/mailru/easyjson and
// github.com/pquerna/ffjson.
func Unmarshal(data []byte, v any) error {
	fastUnmarshaller, ok := v.(easyjson.Unmarshaler)
	if ok {
		return easyjson.Unmarshal(data, fastUnmarshaller)
	}

	return ffjson.Unmarshal(data, v)
}

// UnmarshallWithContext decodes a JSON byte slice into a destination value
// using a context-aware Decoder.
// It follows the same helper shape as Unmarshal, but routes the read through
// the package's context-aware streaming helpers.
func UnmarshallWithContext(ctx context.Context, data []byte, v any) error {
	return NewDecoder(ctx, bytes.NewReader(data)).Decode(v)
}

// ToYAML converts JSON data to YAML.
func ToYAML(rawJSON []byte) (yaml []byte, err error) {
	yaml, err = sigsyaml.JSONToYAML(rawJSON)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed converting JSON to YAML")
	}
	return
}

// IsJSON returns true if the extension is a JSON file
func IsJSON(extension string) bool {
	return collection.In(JSONExtensions, extension, collection.StringCleanCaseInsensitiveMatch)
}
