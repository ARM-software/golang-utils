/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build darwin

package platform

import "fmt"

func determineDefaultHomeDirectory(username string) (string, error) {
	return fmt.Sprintf("/Users/%v", username), nil
}
