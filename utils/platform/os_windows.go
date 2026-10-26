/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

package platform

import "golang.org/x/sys/windows/registry"

func expandFromEnvironment(s string) string {
	expanded, err := registry.ExpandString(s)
	if err == nil {
		return expanded
	}
	return s
}
