/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Distributed lock using lock files https://fileinfo.com/extension/lock
package filesystem

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const LockFilePrefix = "lockfile"

// RemoteLockFile describes a distributed lock using only the file system.
// The locking mechanism is performed using directories and the atomic function `mkdir`.
// A major issue of distributed locks is the presence of stale locks due to many factors such as the loss of the holder of a lock for various reasons.
// To mitigate this problem, a "heart bit" file is modified regularly by the lock holder in order to specify the holder is still alive and the lock still valid.
type RemoteLockFile struct {
	id                   string
	prefix               string
	path                 string
	timeBetweenLockTries time.Duration
	fs                   *VFS
	lockHeartBeatPeriod  time.Duration
	cancelStore          *parallelisation.CancelFunctionStore
	overrideStaleLock    bool
}

// NewGenericRemoteLockFile creates a new remote lock using the file system.
func NewGenericRemoteLockFile(fs *VFS, lockID string, dirPath string, overrideStaleLock bool) ILock {
	return &RemoteLockFile{
		id:                   lockID,
		prefix:               LockFilePrefix,
		path:                 dirPath,
		timeBetweenLockTries: 10 * time.Millisecond,
		fs:                   fs,
		lockHeartBeatPeriod:  50 * time.Millisecond,
		cancelStore:          parallelisation.NewCancelFunctionsStore(),
		overrideStaleLock:    overrideStaleLock,
	}
}

// NewRemoteLockFile creates a new remote lock using the file system.
// lockID Id for the lock.
// dirPath path where the lock should be applied to.
func NewRemoteLockFile(fs *VFS, lockID string, dirPath string) ILock {
	return NewGenericRemoteLockFile(fs, lockID, dirPath, false)
}

func heartBeat(ctx context.Context, fs FS, period time.Duration, filepath string) {
	for {
		if err := parallelisation.DetermineContextError(ctx); err != nil {
			return
		}
		now := time.Now()
		_ = fs.WriteFile(filepath, []byte(fmt.Sprintf("alive @ %v", now)), 0775)
		// FIXME: this is to overcome the problem found with different filesystems which do not update modTime on file change.
		// e.g. https://github.com/spf13/afero/issues/297
		_ = fs.Chtimes(filepath, now, now)
		// sleeping until next heart beat
		parallelisation.SleepWithContext(ctx, period-time.Millisecond)
	}
}
func (l *RemoteLockFile) lockPath() string {
	return FilePathJoin(l.fs, l.path, fmt.Sprintf("%v-%v", strings.TrimSpace(l.prefix), strings.TrimSpace(l.id)))
}

// IsStale checks whether the lock is stale (i.e. no heart beat detected) or not.
func (l *RemoteLockFile) IsStale() bool {
	lockPath := l.lockPath()
	heartBeatFiles, err := l.fs.Ls(lockPath)
	if err != nil {
		return false
	}
	if len(heartBeatFiles) == 0 {
		// if directory exists but no files are present, then it could be that the directory has been created
		// but that the heart beat file hasn't yet. Therefore we check the age of the directory and deduce whether
		// it is stale or not.
		dirInfo, err := l.fs.StatTimes(lockPath)
		if err != nil {
			return false
		}
		return isStale(dirInfo, l.lockHeartBeatPeriod)
	}
	return areHeartBeatFilesAllStale(l.fs, lockPath, heartBeatFiles, l.lockHeartBeatPeriod)
}

func areHeartBeatFilesAllStale(fs *VFS, lockPath string, heartBeatFiles []string, lockHeartBeatPeriod time.Duration) bool {
	staleFiles := []bool{}
	for i := range heartBeatFiles {
		heartBeat := FilePathJoin(fs, lockPath, heartBeatFiles[i]) // there should only be one file in the directory
		// check the time since the heart beat was last modified.
		// if this is less than that beat period then the lock is alive
		info, err := fs.StatTimes(heartBeat)
		isStaleB := false
		if err == nil {
			isStaleB = isStale(info, lockHeartBeatPeriod)
		}
		staleFiles = append(staleFiles, isStaleB)
	}
	return collection.All(staleFiles)
}

func isStale(filetime FileTimeInfo, beatPeriod time.Duration) bool {
	if filetime == nil {
		return false
	}
	return time.Since(filetime.ModTime()).Milliseconds() > 2*beatPeriod.Milliseconds()
}

func (l *RemoteLockFile) ReleaseIfStale(ctx context.Context) error {
	if l.IsStale() {
		return l.Unlock(ctx)
	}
	return nil
}

// TryLock attempts to lock the lock straight away.
func (l *RemoteLockFile) TryLock(ctx context.Context) (err error) {
	if err := parallelisation.DetermineContextError(ctx); err != nil {
		return err
	}

	lockPath := l.lockPath()
	// create directory as lock
	err = l.fs.vfs.Mkdir(lockPath, 0755)
	if commonerrors.Any(ConvertFileSystemError(err), commonerrors.ErrExists) {
		if l.IsStale() {
			if l.overrideStaleLock {
				_ = l.ReleaseIfStale(ctx)
				err = l.TryLock(ctx)
				return err
			}
			return commonerrors.ErrStaleLock
		}
		return commonerrors.ErrLocked
	}
	if err != nil {
		return
	}

	// FIXME: the following is to overcome the problem found with different filesystems which do not update modTime on directory creation.
	// e.g. https://github.com/spf13/afero/issues/297
	now := time.Now()
	_ = l.fs.Chtimes(lockPath, now, now)
	// create a heart beat file that will be updated whilst the lock is active
	// there will be a context for cancelling update status when unlock is called
	// the status file will update the file (modtime) until told to cancel through ctx
	heartBeatFilePath := l.heartBeatFile(lockPath)
	subctx, cancelFunc := context.WithCancel(ctx)
	l.cancelStore.RegisterCancelFunction(cancelFunc)
	go heartBeat(subctx, l.fs, l.lockHeartBeatPeriod, heartBeatFilePath)
	return nil
}

func (l *RemoteLockFile) heartBeatFile(lockPath string) string {
	return FilePathJoin(l.fs, lockPath, fmt.Sprintf("%v.lock", l.id))
}

// Lock locks the lock. This call will block until the lock is available.
func (l *RemoteLockFile) Lock(ctx context.Context) error {
	for {
		if err := parallelisation.DetermineContextError(ctx); err != nil {
			return err
		}
		if err := l.TryLock(ctx); err != nil {
			if err == commonerrors.ErrLocked {
				waitCtx, cancel := context.WithTimeout(ctx, l.timeBetweenLockTries)
				<-waitCtx.Done()
				cancel()
			} else {
				return err
			}
		} else {
			return nil
		}
	}
}

// LockWithTimeout tries to lock the lock until the timeout expires
func (l *RemoteLockFile) LockWithTimeout(ctx context.Context, timeout time.Duration) error {
	if err := parallelisation.DetermineContextError(ctx); err != nil {
		return err
	}
	return parallelisation.RunActionWithTimeoutAndCancelStore(ctx, timeout, l.cancelStore, l.Lock)
}

// Unlock unlocks the lock
func (l *RemoteLockFile) Unlock(ctx context.Context) error {
	l.cancelStore.Cancel()
	return retry.Do(
		func() error {
			err := l.fs.Rm(l.lockPath())
			if err != nil {
				return commonerrors.Newf(err, "cannot unlock lock [%v]", l.id)
			}
			if l.fs.Exists(l.lockPath()) {
				return commonerrors.Newf(commonerrors.ErrLocked, "cannot unlock lock [%v]", l.id)
			}
			return nil
		},
		retry.MaxJitter(25*time.Millisecond),
		retry.DelayType(retry.RandomDelay),
		retry.Attempts(10),
		retry.Context(ctx),
	)
}

// MakeStale is mostly useful for testing purposes and tries to mock locks going stale.
func (l *RemoteLockFile) MakeStale(ctx context.Context) error {
	l.cancelStore.Cancel()
	parallelisation.SleepWithContext(ctx, l.lockHeartBeatPeriod+time.Millisecond)
	lockPath := l.lockPath()
	filePath := l.heartBeatFile(lockPath)
	newTime := time.Now().Add(-1 * (l.lockHeartBeatPeriod + time.Millisecond))
	return retry.Do(
		func() error {
			if !l.fs.Exists(lockPath) {
				return nil
			}
			if l.fs.Exists(filePath) {
				_ = l.fs.Chtimes(filePath, newTime, newTime)
			} else {
				_ = l.fs.Chtimes(lockPath, newTime, newTime)
			}
			if !l.IsStale() {
				return commonerrors.Newf(commonerrors.ErrConflict, "cannot make lock [%v] stale", l.id)
			}
			return nil
		},
		retry.MaxJitter(l.lockHeartBeatPeriod),
		retry.DelayType(retry.RandomDelay),
		retry.Attempts(10),
		retry.Context(ctx),
	)
}
