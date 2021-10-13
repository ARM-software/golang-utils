/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/hashing"
)

type fileHashing struct {
	algo hashing.IHash
}

func (h *fileHashing) Calculate(f File) (string, error) {
	if f == nil {
		return "", commonerrors.ErrUndefined
	}
	return h.algo.Calculate(f)
}

func (h *fileHashing) GetType() string {
	return h.algo.GetType()
}

func (h *fileHashing) CalculateFile(fs FS, path string) (string, error) {
	ok, err := fs.IsFile(path)
	if err != nil || !ok {
		if err != nil {
			return "", err
		}
		err = fmt.Errorf("not a file [%v]: %w", path, commonerrors.ErrInvalid)
		return "", err
	}
	f, err := fs.GenericOpen(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	return h.Calculate(f)
}

func NewFileHash(hashType string) (IFileHash, error) {
	algo, err := hashing.NewHashingAlgorithm(hashType)
	if err != nil {
		return nil, err
	}
	return &fileHashing{
		algo: algo,
	}, nil
}
