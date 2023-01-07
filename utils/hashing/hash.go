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
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/spaolacci/murmur3"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
	strings2 "github.com/ARM-software/golang-utils/utils/strings"
)

const (
	HashMd5    = "MD5"
	HashSha256 = "SHA256"
	HashSha1   = "SHA1"
	HashMurmur = "Murmur"
	HashXXHash = "xxhash" //https://github.com/OneOfOne/xxhash
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

func NewHashingAlgorithm(htype string) (IHash, error) {
	var hash hash.Hash
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
	}

	if hash == nil {
		return nil, commonerrors.ErrNotFound
	}
	return &hashingAlgo{
		Hash: hash,
		Type: htype,
	}, nil
}

func CalculateMD5Hash(text string) string {
	return CalculateHash(text, HashMd5)
}

func CalculateHash(text, htype string) string {
	hashing, err := NewHashingAlgorithm(htype)
	if err != nil {
		return ""
	}
	hash, err := hashing.Calculate(strings.NewReader(text))
	if err != nil {
		return ""
	}
	return hash
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
