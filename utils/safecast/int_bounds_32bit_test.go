/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build 386 || arm || mips || mipsle || wasm

package safecast

import "math"

const (
	testArchMaxInt  = math.MaxInt32
	testArchMinInt  = math.MinInt32
	testArchMaxUint = uint(math.MaxUint32)
)

func expectedIntFromAboveInt32(_ int64) int {
	return testArchMaxInt
}

func expectedIntFromBelowInt32(_ int64) int {
	return testArchMinInt
}

func expectedUintFromAboveUint32(_ uint64) uint {
	return testArchMaxUint
}
