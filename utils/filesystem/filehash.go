/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/hashing"
)

type fileHashing struct {
	algo hashing.IHash
}

func (h *fileHashing) CalculateWithContext(ctx context.Context, f File) (string, error) {
	if f == nil {
		return "", commonerrors.ErrUndefined
	}
	return h.algo.CalculateWithContext(ctx, f)
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

func (h *fileHashing) CalculateFileWithContext(ctx context.Context, fs FS, path string) (string, error) {
	return h.calculateFile(fs, path, func(m *fileHashing, f File) (string, error) { return m.CalculateWithContext(ctx, f) })
}

func (h *fileHashing) CalculateFile(fs FS, path string) (string, error) {
	return h.calculateFile(fs, path, func(m *fileHashing, f File) (string, error) { return m.Calculate(f) })
}

func (h *fileHashing) calculateFile(fs FS, path string, hashFunc func(h *fileHashing, f File) (string, error)) (string, error) {
	ok, err := fs.IsFile(path)
	if err != nil || !ok {
		if err != nil {
			return "", err
		}
		err = commonerrors.Newf(commonerrors.ErrInvalid, "not a file [%v]", path)
		return "", err
	}
	f, err := fs.GenericOpen(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	return hashFunc(h, f)
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
