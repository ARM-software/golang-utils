/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialization_test //nolint:misspell // Verify historical public import paths.

import (
	"testing"

	"github.com/stretchr/testify/require"

	britishjson "github.com/ARM-software/golang-utils/utils/serialisation/json"
	britishmaps "github.com/ARM-software/golang-utils/utils/serialisation/maps"
	britishxml "github.com/ARM-software/golang-utils/utils/serialisation/xml"
	britishyaml "github.com/ARM-software/golang-utils/utils/serialisation/yaml"
	legacy "github.com/ARM-software/golang-utils/utils/serialization"          //nolint:misspell
	legacyjson "github.com/ARM-software/golang-utils/utils/serialization/json" //nolint:misspell,staticcheck // Exercise the historical import path.
	legacymaps "github.com/ARM-software/golang-utils/utils/serialization/maps" //nolint:misspell,staticcheck // Exercise the historical import path.
	legacyxml "github.com/ARM-software/golang-utils/utils/serialization/xml"   //nolint:misspell,staticcheck // Exercise the historical import path.
	legacyyaml "github.com/ARM-software/golang-utils/utils/serialization/yaml" //nolint:misspell,staticcheck // Exercise the historical import path.
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
		require.NoError(t, britishxml.UnmarshallXML(content, &newValue))
		require.Equal(t, document{Name: "preserved"}, rootValue)
		require.Equal(t, newValue, oldValue)
	})

	t.Run("JSON and YAML", func(t *testing.T) {
		value := struct {
			Name string `json:"name" yaml:"name"`
		}{Name: "preserved"}
		oldJSON, err := legacyjson.Marshal(value)
		require.NoError(t, err)
		newJSON, err := britishjson.Marshal(value)
		require.NoError(t, err)
		require.JSONEq(t, string(newJSON), string(oldJSON))

		oldYAML, err := legacyyaml.Marshal(value)
		require.NoError(t, err)
		newYAML, err := britishyaml.Marshal(value)
		require.NoError(t, err)
		require.Equal(t, string(newYAML), string(oldYAML))
	})

	t.Run("map helpers", func(t *testing.T) {
		value := struct{ Name string }{Name: "preserved"}
		oldMap, err := legacymaps.ToMap(&value)
		require.NoError(t, err)
		newMap, err := britishmaps.ToMap(&value)
		require.NoError(t, err)
		require.Equal(t, newMap, oldMap)
	})
}
