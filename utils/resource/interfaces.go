/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package resource resource defines utilities about generic resources.
package resource

import (
	"fmt"
	"io"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE ICloseableResource

// ICloseableResource defines a resource which must be closed after use e.g. an open file.
type ICloseableResource interface {
	io.Closer
	fmt.Stringer
	IsClosed() bool
}
