/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"context"
	"os"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/platform"
)

// ExtendedOsFs extends afero.OsFs and is a Fs implementation that uses functions provided by the os package.
type ExtendedOsFs struct {
	afero.OsFs
}

func (c *ExtendedOsFs) ChownIfPossible(name string, uid int, gid int) error {
	return ConvertFileSystemError(c.OsFs.Chown(name, uid, gid))
}

func (c *ExtendedOsFs) LinkIfPossible(oldname, newname string) (err error) {
	return ConvertFileSystemError(os.Link(oldname, newname))
}

func (c *ExtendedOsFs) ForceRemoveIfPossible(path string) error {
	return ConvertFileSystemError(platform.RemoveWithPrivileges(context.Background(), path))
}

func NewExtendedOsFs() afero.Fs {
	return &ExtendedOsFs{}
}
