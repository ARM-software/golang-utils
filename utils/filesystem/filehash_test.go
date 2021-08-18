/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/hashing"
)

func TestFileHash(t *testing.T) {
	testInDir := "testdata"
	testFilePath := filepath.Join(testInDir, "1KB.bin")
	fs := NewFs(StandardFS)
	tests := []struct {
		Type string
		Hash string
	}{
		//values given by https://www.pelock.com/products/hash-calculator
		{
			Type: hashing.HashMd5,
			Hash: "CBF17A648BBBCDD7AB591784E96F85C7",
		},
		{
			Type: hashing.HashSha1,
			Hash: "34D3F2C19C7846F3BD6B817EB4AD85DA6CFF5B0F",
		},
		{
			Type: hashing.HashSha256,
			Hash: "1920FE73F1DA83A6551B0D2404ED0BD6EC7BB24D941B99A84ADA7D37C67A0527",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Hash %v", test.Type), func(t *testing.T) {
			hasher, err := NewFileHash(test.Type)
			require.Nil(t, err)
			hash, err := hasher.CalculateFile(fs, testFilePath)
			require.Nil(t, err)
			assert.Equal(t, strings.ToLower(test.Hash), strings.ToLower(hash))
		})

	}
}

func TestFileHash2(t *testing.T) {
	testInDir := "testdata"
	fs := NewFs(StandardFS)
	hasher, err := NewFileHash(hashing.HashSha256)
	require.Nil(t, err)
	tests := []struct {
		Path string
		Hash string
	}{
		//values given by https://www.pelock.com/products/hash-calculator
		{
			Path: "1KB.bin",
			Hash: "1920FE73F1DA83A6551B0D2404ED0BD6EC7BB24D941B99A84ADA7D37C67A0527",
		},
		{
			Path: "5MB.zip",
			Hash: "C0DE104C1E68625629646025D15A6129A2B4B6496CD9CEACD7F7B5078E1849BA",
		},
		{
			Path: "testunzip.zip",
			Hash: "F5CAF5513A796443914288CF2A2586F3CCE1730AADA8DD397822FFECBBA5BA26",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Hash %v", test.Path), func(t *testing.T) {
			hash, err := hasher.CalculateFile(fs, filepath.Join(testInDir, test.Path))
			require.Nil(t, err)
			assert.Equal(t, strings.ToLower(test.Hash), strings.ToLower(hash))
		})

	}
}
