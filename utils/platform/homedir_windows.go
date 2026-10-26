/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

package platform

import (
	"fmt"
	"os"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func determineDefaultHomeDirectory(username string) (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	if drive == "" {
		return "", fmt.Errorf("%w: cannot determine the default home drive", commonerrors.ErrUnexpected)
	}
	return fmt.Sprintf("%v\\Users\\%v", drive, username), nil
}
