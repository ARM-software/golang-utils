/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package idgen

import (
	"fmt"

	"github.com/gofrs/uuid/v5"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// Generates a UUID.
func GenerateUUID4() (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("%w: failed generating uuid: %v", commonerrors.ErrUnexpected, err.Error())
	}
	return uuid.String(), nil
}

func IsValidUUID(u string) bool {
	_, err := uuid.FromString(u)
	return err == nil
}
