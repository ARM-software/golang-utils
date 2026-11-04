/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x

package safecast

func expectedIntFromAboveInt32(value int64) int {
	return int(value)
}

func expectedIntFromBelowInt32(value int64) int {
	return int(value)
}

func expectedUintFromAboveUint32(value uint64) uint {
	return uint(value)
}
