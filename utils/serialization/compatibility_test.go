/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialization_test //nolint:misspell // Verify historical public import paths.

import (
	"testing"

	"github.com/stretchr/testify/require"

	preferredfrontmatter "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
	preferredjson "github.com/ARM-software/golang-utils/utils/serialisation/json"
	preferredeasyjson "github.com/ARM-software/golang-utils/utils/serialisation/json/jsontest/easyjson"
	preferredmaps "github.com/ARM-software/golang-utils/utils/serialisation/maps"
	preferredmapstest "github.com/ARM-software/golang-utils/utils/serialisation/maps/testing"
	preferredxml "github.com/ARM-software/golang-utils/utils/serialisation/xml"
	preferredyaml "github.com/ARM-software/golang-utils/utils/serialisation/yaml"
	legacy "github.com/ARM-software/golang-utils/utils/serialization"                                //nolint:misspell
	legacyfrontmatter "github.com/ARM-software/golang-utils/utils/serialization/frontmatter"         //nolint:misspell,staticcheck // Verify the historical alias.
	legacyjson "github.com/ARM-software/golang-utils/utils/serialization/json"                       //nolint:misspell,staticcheck // Exercise the historical import path.
	legacyeasyjson "github.com/ARM-software/golang-utils/utils/serialization/json/jsontest/easyjson" //nolint:misspell // Verify the historical alias.
	legacymaps "github.com/ARM-software/golang-utils/utils/serialization/maps"                       //nolint:misspell,staticcheck // Exercise the historical import path.
	legacymapstest "github.com/ARM-software/golang-utils/utils/serialization/maps/testing"           //nolint:misspell,staticcheck // Verify the historical alias.
	legacyxml "github.com/ARM-software/golang-utils/utils/serialization/xml"                         //nolint:misspell,staticcheck // Exercise the historical import path.
	legacyyaml "github.com/ARM-software/golang-utils/utils/serialization/yaml"                       //nolint:misspell,staticcheck // Exercise the historical import path.
)

// These assignments require the historical exported types to be identical to
// their canonical equivalents, rather than merely structurally similar.
var (
	_ *preferredfrontmatter.Format            = (*legacyfrontmatter.Format)(nil)
	_ preferredfrontmatter.FormatOption       = legacyfrontmatter.WithStart("---")
	_ preferredfrontmatter.IFrontMatterParser = (legacyfrontmatter.IFrontMatterParser)(nil)
	_ *preferredjson.Decoder                  = (*legacyjson.Decoder)(nil)
	_ *preferredyaml.Encoder                  = (*legacyyaml.Encoder)(nil)
	_ preferredmapstest.TestEnumWithUnmarshal = legacymapstest.TestEnumWithUnmarshal0
	_ *preferredeasyjson.TestingStruct        = (*legacyeasyjson.TestingStruct)(nil)
)

func TestHistoricalSerialisationImports(t *testing.T) {
	t.Run("XML root and subpackage", func(t *testing.T) {
		type document struct {
			Name string `xml:"name,attr"`
		}
		content := []byte(`<document name="preserved"/>`)
		var rootValue, oldValue, newValue document
		require.NoError(t, legacy.UnmarshallXML(content, &rootValue))
		require.NoError(t, legacyxml.UnmarshallXML(content, &oldValue))
		require.NoError(t, preferredxml.UnmarshallXML(content, &newValue))
		require.Equal(t, document{Name: "preserved"}, rootValue)
		require.Equal(t, newValue, oldValue)
	})

	t.Run("JSON and YAML", func(t *testing.T) {
		value := struct {
			Name string `json:"name" yaml:"name"`
		}{Name: "preserved"}
		oldJSON, err := legacyjson.Marshal(value)
		require.NoError(t, err)
		newJSON, err := preferredjson.Marshal(value)
		require.NoError(t, err)
		require.JSONEq(t, string(newJSON), string(oldJSON))

		oldYAML, err := legacyyaml.Marshal(value)
		require.NoError(t, err)
		newYAML, err := preferredyaml.Marshal(value)
		require.NoError(t, err)
		require.Equal(t, string(newYAML), string(oldYAML))
	})

	t.Run("map helpers", func(t *testing.T) {
		value := struct{ Name string }{Name: "preserved"}
		oldMap, err := legacymaps.ToMap(&value)
		require.NoError(t, err)
		newMap, err := preferredmaps.ToMap(&value)
		require.NoError(t, err)
		require.Equal(t, newMap, oldMap)
	})
}
