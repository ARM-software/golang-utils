/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package idgen

import (
	"github.com/gofrs/uuid/v5"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// Generates a UUID.
func GenerateUUID4() (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed generating uuid")
	}
	return uuid.String(), nil
}

func IsValidUUID(u string) bool {
	_, err := uuid.FromString(u)
	return err == nil
}
