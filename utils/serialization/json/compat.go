/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package json

import (
	"context"
	"io"

	serialisation "github.com/ARM-software/golang-utils/utils/serialisation/json"
)

// Decoder decodes successive JSON values from a reader.
type Decoder = serialisation.Decoder

// Encoder writes successive JSON values to a writer.
type Encoder = serialisation.Encoder

// JSONExtensions lists the file extensions considered JSON files.
var JSONExtensions = serialisation.JSONExtensions

// IsValidJSON validates JSON strings and byte slices.
var IsValidJSON = serialisation.IsValidJSON

// NewDecoder creates a context-aware JSON decoder.
func NewDecoder(ctx context.Context, r io.Reader) *Decoder { return serialisation.NewDecoder(ctx, r) }

// NewEncoder creates a context-aware JSON encoder.
func NewEncoder(ctx context.Context, w io.Writer) *Encoder { return serialisation.NewEncoder(ctx, w) }

// Marshal encodes a value to JSON.
func Marshal(v any) ([]byte, error) { return serialisation.Marshal(v) }

// MarshalWithContext encodes a value to JSON with a context-aware encoder.
func MarshalWithContext(ctx context.Context, v any) ([]byte, error) {
	return serialisation.MarshalWithContext(ctx, v)
}

// Unmarshal decodes JSON into a destination value.
func Unmarshal(data []byte, v any) error { return serialisation.Unmarshal(data, v) }

// UnmarshallWithContext decodes JSON with a context-aware decoder.
func UnmarshallWithContext(ctx context.Context, data []byte, v any) error {
	return serialisation.UnmarshallWithContext(ctx, data, v)
}

// ToYAML converts JSON data to YAML.
func ToYAML(rawJSON []byte) ([]byte, error) { return serialisation.ToYAML(rawJSON) }

// IsJSON reports whether an extension belongs to a JSON file.
func IsJSON(extension string) bool { return serialisation.IsJSON(extension) }
