/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logrimp defines some common logr implementation
package logrimp

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

// NewStdOutLogr returns a logger to standard out.
// See https://github.com/go-logr/logr/blob/ff91da8dc418a9e36998931ed4ab10b71833a368/example_test.go#L27
func NewStdOutLogr() logr.Logger {
	return funcr.New(func(prefix, args string) {
		if prefix != "" {
			fmt.Printf("%s: %s\n", prefix, args)
		} else {
			fmt.Println(args)
		}
	}, funcr.Options{})
}
