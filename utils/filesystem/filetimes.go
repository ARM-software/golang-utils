/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"os"
	"time"

	fileTimes "github.com/djherbis/times"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func DetermineFileTimes(info os.FileInfo) (times FileTimeInfo, err error) {
	if info == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "no file information defined")
		return
	}
	if info.Sys() == nil {
		times = newDefaultTimeInfo(info)
	} else {
		times = &genericTimeInfo{fileTimes.Get(info)}
	}
	return
}

type defaultTimeInfo struct {
	modTime time.Time
}

func (i *defaultTimeInfo) ModTime() time.Time {
	return i.modTime
}
func (i *defaultTimeInfo) AccessTime() time.Time {
	return time.Now()
}
func (i *defaultTimeInfo) ChangeTime() time.Time {
	return time.Now()
}
func (i *defaultTimeInfo) BirthTime() time.Time {
	return time.Now()
}
func (i *defaultTimeInfo) HasChangeTime() bool {
	return false
}
func (i *defaultTimeInfo) HasBirthTime() bool {
	return false
}

func (i *defaultTimeInfo) HasAccessTime() bool {
	return false
}

func newDefaultTimeInfo(f os.FileInfo) (info *defaultTimeInfo) {
	info = &defaultTimeInfo{}
	if f != nil {
		info.modTime = f.ModTime()
	}
	return
}

type genericTimeInfo struct {
	fileTimes.Timespec
}

func (i *genericTimeInfo) HasAccessTime() bool {
	return false
}
