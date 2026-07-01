/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package collection provides utilities for working with slices, maps, and
// sequences.
//
// The package mixes eager helpers for slices with lazy helpers built on top of
// `iter.Seq`, so the same collection-oriented operations can be applied to both
// in-memory data and streamed values.
//
// The overall style is primarily driven by the Go standard library and the
// needs of this repository, with some influence from helper libraries such as:
//   - `samber/lo`: https://github.com/samber/lo
//   - `samber/mo`: https://github.com/samber/mo
//
// Where possible, helpers should complement rather than duplicate what is
// already available in the standard `slices` package.
package collection
