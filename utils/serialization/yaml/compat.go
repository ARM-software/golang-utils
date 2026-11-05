/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package yaml

import (
	"context"
	"io"

	serialisation "github.com/ARM-software/golang-utils/utils/serialisation/yaml"
)

// Decoder decodes YAML from a reader.
type Decoder = serialisation.Decoder

// Encoder writes YAML values to a writer.
type Encoder = serialisation.Encoder

// YAMLExtensions lists the file extensions considered YAML files.
var YAMLExtensions = serialisation.YAMLExtensions

// IsValidYAML validates YAML strings and byte slices.
var IsValidYAML = serialisation.IsValidYAML

// NewDecoder creates a context-aware YAML decoder.
func NewDecoder(ctx context.Context, r io.Reader) *Decoder { return serialisation.NewDecoder(ctx, r) }

// NewEncoder creates a context-aware YAML encoder.
func NewEncoder(ctx context.Context, w io.Writer) *Encoder { return serialisation.NewEncoder(ctx, w) }

// ToJSON converts YAML data to JSON.
func ToJSON(data []byte) ([]byte, error) { return serialisation.ToJSON(data) }

// Marshal encodes a value to YAML.
func Marshal(v any) ([]byte, error) { return serialisation.Marshal(v) }

// MarshalWithContext encodes a value to YAML with a context-aware encoder.
func MarshalWithContext(ctx context.Context, v any) ([]byte, error) {
	return serialisation.MarshalWithContext(ctx, v)
}

// Unmarshal decodes YAML into a destination value.
func Unmarshal(data []byte, v any) error { return serialisation.Unmarshal(data, v) }

// UnmarshallWithContext decodes YAML with a context-aware decoder.
func UnmarshallWithContext(ctx context.Context, data []byte, v any) error {
	return serialisation.UnmarshallWithContext(ctx, data, v)
}

// IsYAMLFile reports whether an extension belongs to a YAML file.
func IsYAMLFile(extension string) bool { return serialisation.IsYAMLFile(extension) }
