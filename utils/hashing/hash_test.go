/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package hashing

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

func TestHasher(t *testing.T) {
	// values given by https://md5calc.com/hash/md5/test
	hasher, err := NewHashingAlgorithm(HashMd5)
	require.Nil(t, err)
	testCases := []struct {
		Input string
		Hash  string
	}{{
		Input: "test",
		Hash:  "098f6bcd4621d373cade4e832627b4f6",
	}, {
		Input: "CMSIS",
		Hash:  "c61d595888f85f6d30e99ef6cacfcb7d",
	}}
	for _, testCase := range testCases {
		hash, err := hasher.Calculate(strings.NewReader(testCase.Input))
		require.Nil(t, err)
		assert.Equal(t, testCase.Hash, hash)
	}
}

func TestMd5(t *testing.T) {
	// values given by https://md5calc.com/hash/md5/test
	testCases := []struct {
		Input string
		Hash  string
	}{{
		Input: "test",
		Hash:  "098f6bcd4621d373cade4e832627b4f6",
	}, {
		Input: "CMSIS",
		Hash:  "c61d595888f85f6d30e99ef6cacfcb7d",
	}}
	for _, testCase := range testCases {
		assert.Equal(t, testCase.Hash, CalculateMD5Hash(testCase.Input))
	}
}

func TestIsLikelyHexHashString(t *testing.T) {
	tests := []struct {
		input  string
		isHash bool
	}{
		{
			input:  "",
			isHash: false,
		},
		{
			input:  faker.Word(),
			isHash: false,
		},
		{
			input:  faker.Name(),
			isHash: false,
		},
		{
			input:  faker.Sentence(),
			isHash: false,
		},
		{
			input:  faker.CCNumber(),
			isHash: false,
		},
		{
			input:  faker.UUIDHyphenated(),
			isHash: false,
		},
		{
			input:  faker.IPv4(),
			isHash: false,
		},
		{
			input:  faker.Paragraph(),
			isHash: false,
		},
		{
			input:  "1.0.1",
			isHash: false,
		},
		{
			input:  "v1.0.1",
			isHash: false,
		},
		{
			input:  CalculateMD5Hash(faker.Paragraph()),
			isHash: true,
		},

		{
			input:  CalculateHash(faker.Paragraph(), HashSha256),
			isHash: true,
		},
		{
			input:  "85817ddeed66c3e3805c73dbc7082de2674e349c",
			isHash: true,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%v_isHash(%v)", i, test.input), func(t *testing.T) {
			require.Equal(t, test.isHash, IsLikelyHexHashString(test.input))
		})
	}
}

func TestBespokeHash(t *testing.T) {
	random, err := faker.RandomInt(1, 64, 1)
	require.NoError(t, err)
	size := random[0]
	algo, err := blake2b.New(size, nil)
	require.NoError(t, err)
	hashing, err := NewBespokeHashingAlgorithm(algo)
	require.NoError(t, err)
	hash := CalculateStringHash(hashing, faker.Paragraph())
	require.NotEmpty(t, hash)
	assert.Equal(t, size*2, len(hash))
}
