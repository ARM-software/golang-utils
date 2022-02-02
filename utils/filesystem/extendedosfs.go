/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"os"

	"github.com/spf13/afero"
)

type ExtendedOsFs struct {
	afero.OsFs
}

func (c *ExtendedOsFs) ChownIfPossible(name string, uid int, gid int) error {
	return os.Chown(name, uid, gid)
}

func (c *ExtendedOsFs) LinkIfPossible(oldname, newname string) (err error) {
	return os.Link(oldname, newname)
}

func NewExtendedOsFs() afero.Fs {
	return &ExtendedOsFs{}
}
