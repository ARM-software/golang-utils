/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hashing

import (
	"context"
	"crypto/md5"  //nolint:gosec
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"math"
	"slices"
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/spaolacci/murmur3"
	"golang.org/x/crypto/blake2b"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
	strings2 "github.com/ARM-software/golang-utils/utils/strings"
)

const (
	HashMd5       = "MD5"
	HashSha256    = "SHA256"
	HashSha1      = "SHA1"
	HashMurmur    = "Murmur"
	HashXXHash    = "xxhash"     // https://github.com/OneOfOne/xxhash
	HashBlake2256 = "blake2b256" // https://www.blake2.net/
)

type hashingAlgo struct {
	Hash hash.Hash
	Type string
}

func (h *hashingAlgo) CalculateWithContext(ctx context.Context, r io.Reader) (hashN string, err error) {
	if r == nil {
		err = commonerrors.ErrUndefined
		return
	}
	_, err = safeio.CopyDataWithContext(ctx, r, h.Hash)
	if err != nil {
		return
	}
	hashN = hex.EncodeToString(h.Hash.Sum(nil))
	h.Hash.Reset()
	return
}

func (h *hashingAlgo) Calculate(r io.Reader) (string, error) {
	return h.CalculateWithContext(context.Background(), r)
}

func (h *hashingAlgo) GetType() string {
	return h.Type
}

// NewBespokeHashingAlgorithm defines a bespoke hashing algorithm
func NewBespokeHashingAlgorithm(algorithm hash.Hash) (IHash, error) {
	return newHashingAlgorithm("bespoke", algorithm)
}

func newHashingAlgorithm(htype string, algorithm hash.Hash) (IHash, error) {
	return &hashingAlgo{
		Hash: algorithm,
		Type: htype,
	}, nil
}

// DetermineHashingAlgorithmCanonicalReference determines the hashing algorithm reference from a string.
func DetermineHashingAlgorithmCanonicalReference(name string) (ref string, err error) {
	n := strings.TrimSpace(strings.ReplaceAll(name, "-", ""))
	if reflection.IsEmpty(n) {
		err = commonerrors.UndefinedVariable("algorithm name")
		return
	}
	switch {
	case strings.EqualFold(HashMd5, n):
		ref = HashMd5
	case strings.EqualFold(HashSha1, n):

		ref = HashSha1
	case strings.EqualFold(HashSha256, n):
		ref = HashSha256
	case strings.EqualFold(HashMurmur, n):
		ref = HashMurmur
	case strings.EqualFold(HashXXHash, n):
		ref = HashXXHash
	case strings.EqualFold(HashBlake2256, n):
		ref = HashBlake2256
	default:
		err = commonerrors.New(commonerrors.ErrNotFound, "could not find the corresponding hashing algorithm")
	}
	return
}

// DetermineHashingAlgorithm returns a hashing algorithm based on a string reference. Similar to NewHashingAlgorithm but more flexible.
func DetermineHashingAlgorithm(algorithm string) (IHash, error) {
	htype, err := DetermineHashingAlgorithmCanonicalReference(algorithm)
	if err != nil {
		return nil, err
	}
	return NewHashingAlgorithm(htype)
}

func NewHashingAlgorithm(htype string) (IHash, error) {
	var hash hash.Hash
	var err error
	switch htype {
	case HashMd5:
		hash = md5.New() //nolint:gosec
	case HashSha1:
		hash = sha1.New() //nolint:gosec
	case HashSha256:
		hash = sha256.New()
	case HashMurmur:
		hash = murmur3.New64()
	case HashXXHash:
		hash = xxhash.New64()
	case HashBlake2256:
		hash, err = blake2b.New256(nil)
	}
	if err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed loading the hashing algorithm")
	}

	if hash == nil {
		return nil, commonerrors.New(commonerrors.ErrNotFound, "could not find the corresponding hashing algorithm")
	}
	return newHashingAlgorithm(htype, hash)
}

func CalculateMD5Hash(text string) string {
	return CalculateHash(text, HashMd5)
}

// CalculateStringHash returns the hash of some text using a particular hashing algorithm
func CalculateStringHash(hashingAlgo IHash, text string) string {
	if hashingAlgo == nil {
		return ""
	}
	hash, err := hashingAlgo.Calculate(strings.NewReader(text))
	if err != nil {
		return ""
	}
	return hash
}

// CalculateStringHashWithContext returns the hash of some text using a particular hashing algorithm
func CalculateStringHashWithContext(ctx context.Context, hashingAlgo IHash, text string) string {
	if hashingAlgo == nil {
		return ""
	}
	hash, err := hashingAlgo.CalculateWithContext(ctx, strings.NewReader(text))
	if err != nil {
		return ""
	}
	return hash
}

// CalculateStringHashList returns the hash of a list of strings using a particular hashing algorithm
func CalculateStringHashList(ctx context.Context, hashingAlgo IHash, text ...string) string {
	if hashingAlgo == nil {
		return ""
	}
	if len(text) == 0 {
		return CalculateStringHashWithContext(ctx, hashingAlgo, "")
	}
	slices.Sort(text)
	return CalculateStringHashWithContext(ctx, hashingAlgo, strings.Join(text, " "))
}

// CalculateHash calculates the hash of some text using the requested htype hashing algorithm.
func CalculateHash(text, htype string) string {
	hashing, err := NewHashingAlgorithm(htype)
	if err != nil {
		return ""
	}
	return CalculateStringHash(hashing, text)
}

// CalculateHashWithContext calculates the hash of some text using the requested htype hashing algorithm.
func CalculateHashWithContext(ctx context.Context, text, htype string) string {
	hashing, err := NewHashingAlgorithm(htype)
	if err != nil {
		return ""
	}
	return CalculateStringHashWithContext(ctx, hashing, text)
}

// CalculateHashOfListOfStrings calculates the hash of some text using the requested htype hashing algorithm.
func CalculateHashOfListOfStrings(ctx context.Context, htype string, text ...string) string {
	hashing, err := NewHashingAlgorithm(htype)
	if err != nil {
		return ""
	}
	return CalculateStringHashList(ctx, hashing, text...)
}

// HasLikelyHexHashStringEntropy states whether a string has an entropy which may entail it is a hexadecimal hash
// This is based on the work done by `detect-secrets` https://github.com/Yelp/detect-secrets/blob/2fc0e31f067af98d97ad0f507dac032c9506f667/detect_secrets/plugins/high_entropy_strings.py#L150
func HasLikelyHexHashStringEntropy(str string) bool {
	entropy := strings2.CalculateStringShannonEntropy(str)
	entropy -= 1.2 / math.Log2(float64(len(str)))
	return entropy > 3.0
}

// IsLikelyHexHashString determines whether the string is likely to be a hexadecimal hash or not.
func IsLikelyHexHashString(str string) bool {
	if reflection.IsEmpty(str) {
		return false
	}
	_, err := hex.DecodeString(str)
	if err != nil {
		return false
	}
	return HasLikelyHexHashStringEntropy(str)
}
