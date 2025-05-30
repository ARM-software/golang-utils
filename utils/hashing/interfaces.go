/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package hashing provides utilities for calculating hashes.
package hashing

import (
	"context"
	"io"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IHash

// IHash defines a hashing algorithm.
type IHash interface {
	Calculate(reader io.Reader) (string, error)
	CalculateWithContext(ctx context.Context, reader io.Reader) (string, error)
	GetType() string
}
