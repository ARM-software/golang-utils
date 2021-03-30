package hashing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasher(t *testing.T) {
	//values given by https://md5calc.com/hash/md5/test
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
	//values given by https://md5calc.com/hash/md5/test
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
