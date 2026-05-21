// Package yaml mirrors the repository's serialization/json helper shape for
// YAML data.
//
// The package converts YAML input to JSON before delegating to the JSON helpers,
// which keeps one implementation of the fast, non-reflective JSON decoding
// paths and their context-aware stream handling. On the write path, it encodes
// to JSON first, then converts that JSON to YAML.
//
// YAML parsing support comes from sigs.k8s.io/yaml, which uses yaml/go-yaml
// underneath. In practice that means the package supports most of YAML 1.2
// while preserving some YAML 1.1 behaviour for compatibility, including YAML
// aliases and anchors, as documented at
// https://github.com/yaml/go-yaml#compatibility.
package yaml

import (
	"bytes"
	"context"

	sigsyaml "sigs.k8s.io/yaml"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	jsonserialization "github.com/ARM-software/golang-utils/utils/serialization/json"
)

// ToJSON converts YAML data to JSON.
//
// YAML parsing support comes from sigs.k8s.io/yaml, which uses yaml/go-yaml
// underneath. In practice that means it supports most of YAML 1.2 while
// preserving some YAML 1.1 behaviour for compatibility, as documented at
// https://github.com/yaml/go-yaml#compatibility. That includes support for
// YAML aliases and anchors.
// See also https://github.com/yaml/go-yaml.
func ToJSON(yaml []byte) (json []byte, err error) {
	json, err = sigsyaml.YAMLToJSON(yaml)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed converting YAML to JSON")
	}
	return
}

// Marshal encodes a value to a YAML byte slice.
// It follows the same helper shape as serialization/json.Marshal.
//
// Values are first encoded through the JSON helpers so any generated fast JSON
// serialisers can still be used, then the resulting JSON is converted to YAML.
func Marshal(v any) ([]byte, error) {
	jsonData, err := jsonserialization.Marshal(v)
	if err != nil {
		return nil, err
	}

	return jsonserialization.ToYAML(jsonData)
}

// MarshalWithContext encodes a value to a YAML byte slice using a context-aware
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

// Unmarshal decodes a YAML byte slice into a destination value.
// It follows the same helper shape as serialization/json.Unmarshal.
//
// YAML is converted to JSON first, so the JSON helpers can retain their fast-
// generated decoding paths and shared decoding behaviour.
func Unmarshal(data []byte, v any) error {
	jsonData, err := ToJSON(data)
	if err != nil {
		return err
	}

	return jsonserialization.Unmarshal(jsonData, v)
}

// UnmarshallWithContext decodes a YAML byte slice into a destination value
// using a context-aware Decoder.
// It follows the same helper shape as Unmarshal, but routes the read through
// the package's context-aware streaming helpers.
func UnmarshallWithContext(ctx context.Context, data []byte, v any) error {
	return NewDecoder(ctx, bytes.NewReader(data)).Decode(v)
}
