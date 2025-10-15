/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

var (
	random = rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
)

func TestLockStale(t *testing.T) {
	lockFuncs := []struct {
		LockFunc func(l ILock, ctx context.Context) error
		Name     string
	}{
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				return l.TryLock(ctx)
			},
			Name: "TryLock",
		}, {
			LockFunc: func(l ILock, ctx context.Context) error {
				timeout := 200 * time.Millisecond
				return l.LockWithTimeout(ctx, timeout)
			},
			Name: "LockWithTimeout",
		},
	}
	test := func(t0 *testing.T, ctx context.Context, fs FS, LockFunc func(l ILock, ctx context.Context) error) {
		dirToLock, err := fs.TempDirInTempDir(fmt.Sprintf("test-lock-dir-%v", faker.DomainName()))
		require.Nil(t0, err)
		defer func() { _ = fs.Rm(dirToLock) }()

		lock := fs.NewRemoteLockFile("lock", dirToLock)
		defer func() { _ = lock.Unlock(ctx) }()

		err = lock.Unlock(ctx)
		require.Nil(t0, err)

		err = LockFunc(lock, ctx)
		require.Nil(t0, err)

		assert.False(t0, lock.IsStale())

		time.Sleep(150 * time.Millisecond)

		assert.False(t0, lock.IsStale())

		time.Sleep(150 * time.Millisecond)

		assert.False(t0, lock.IsStale())

		err = lock.MakeStale(ctx)
		require.Nil(t0, err)

		assert.True(t0, lock.IsStale())

		err = lock.Unlock(ctx)
		require.Nil(t0, err)

		err = fs.Rm(dirToLock)
		require.Nil(t0, err)
	}
	for i := range lockFuncs {
		lockFunc := lockFuncs[i]
		for j := range FileSystemTypes {
			fsType := FileSystemTypes[j]
			t.Run(fmt.Sprintf("%v_for_fs_%v_and_%v", t.Name(), fsType, lockFunc.Name), func(t *testing.T) {
				defer goleak.VerifyNone(t)
				fs := NewFs(fsType)
				ctx := context.Background()
				for c := 0; c < 5; c++ {
					test(t, ctx, fs, lockFunc.LockFunc)
				}
			})
		}
	}
}

func TestLockReleaseIfStale(t *testing.T) {
	lockFuncs := []struct {
		LockFunc func(l ILock, ctx context.Context) error
		Name     string
	}{
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				return l.TryLock(ctx)
			},
			Name: "TryLock",
		}, {
			LockFunc: func(l ILock, ctx context.Context) error {
				timeout := 100 * time.Millisecond
				return l.LockWithTimeout(ctx, timeout)
			},
			Name: "LockWithTimeout",
		},
	}
	test := func(t0 *testing.T, ctx context.Context, fs FS, LockFunc func(l ILock, ctx context.Context) error) {
		dirToLock, err := fs.TempDirInTempDir(fmt.Sprintf("test-lock-dir-%v", faker.DomainName()))
		require.Nil(t0, err)
		defer func() { _ = fs.Rm(dirToLock) }()

		lock := fs.NewRemoteLockFile("lock", dirToLock)
		defer func() { _ = lock.Unlock(ctx) }()

		err = lock.Unlock(ctx)
		require.Nil(t0, err)

		err = LockFunc(lock, ctx)
		require.Nil(t0, err)

		err = lock.MakeStale(ctx)
		require.Nil(t0, err)

		assert.True(t0, lock.IsStale())

		err = LockFunc(lock, ctx)
		require.NotNil(t0, err)

		err = lock.ReleaseIfStale(ctx)
		require.Nil(t0, err)

		err = LockFunc(lock, ctx)
		require.Nil(t0, err)

		err = lock.Unlock(ctx)
		require.Nil(t0, err)

		err = fs.Rm(dirToLock)
		require.Nil(t0, err)
	}
	for i := range lockFuncs {
		lockFunc := lockFuncs[i]
		for j := range FileSystemTypes {
			fsType := FileSystemTypes[j]
			t.Run(fmt.Sprintf("%v_for_fs_%v_and_%v", t.Name(), fsType, lockFunc.Name), func(t *testing.T) {
				defer goleak.VerifyNone(t)
				fs := NewFs(fsType)
				ctx := context.Background()
				for c := 0; c < 5; c++ {
					test(t, ctx, fs, lockFunc.LockFunc)
				}
			})
		}
	}
}

func TestLockSimpleSequential(t *testing.T) { // Several lock/unlock sequences performed on a same lock
	lockFuncs := []struct {
		LockFunc      func(l ILock, ctx context.Context) error
		Name          string
		ExpectedError error
	}{
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				return l.TryLock(ctx)
			},
			Name:          "TryLock",
			ExpectedError: commonerrors.ErrLocked,
		},
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				return l.LockWithTimeout(ctx, 200*time.Millisecond)
			},
			Name:          "LockWithTimeout",
			ExpectedError: commonerrors.ErrTimeout,
		},
	}
	test := func(t0 *testing.T, ctx context.Context, fs FS, LockFunc func(l ILock, ctx context.Context) error, expectedError error) {
		dirToLock, err := fs.TempDirInTempDir(fmt.Sprintf("test-lock-dir-%v", faker.DomainName()))
		require.Nil(t0, err)
		defer func() { _ = fs.Rm(dirToLock) }()
		id := "lock"
		Lock := fs.NewRemoteLockFile(id, dirToLock)
		defer func() { _ = Lock.Unlock(ctx) }()

		for c := 0; c < 20; c++ {
			err = Lock.Unlock(ctx)
			require.Nil(t0, err)

			err = LockFunc(Lock, ctx)
			// FIXME it was noticed that there could be some race conditions happening in the in-memory file system when dealing with concurrency
			// see https://github.com/spf13/afero/issues/298
			if fs.GetType() != InMemoryFS {
				require.Nil(t0, err)
			}

			err = Lock.Unlock(ctx)
			require.Nil(t0, err)
		}

		err = fs.Rm(dirToLock)
		require.Nil(t0, err)
	}
	for i := range lockFuncs {
		lockFunc := lockFuncs[i]
		for j := range FileSystemTypes {
			fsType := FileSystemTypes[j]
			t.Run(fmt.Sprintf("%v_for_fs_%v_and_%v", t.Name(), fsType, lockFunc.Name), func(t *testing.T) {
				defer goleak.VerifyNone(t)
				fs := NewFs(fsType)
				ctx := context.Background()
				for c := 0; c < 5; c++ {
					test(t, ctx, fs, lockFunc.LockFunc, lockFunc.ExpectedError)
				}
			})
		}
	}
}

func TestLockSequential(t *testing.T) {
	lockFuncs := []struct {
		LockFunc      func(l ILock, ctx context.Context) error
		Name          string
		ExpectedError error
	}{
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				var err error
				for i := 0; i < 10; i++ {
					err = l.TryLock(ctx)
					if err == nil {
						return err
					}
					time.Sleep(time.Duration(random.Intn(15)) * time.Millisecond) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				}
				return err
			},
			Name:          "TryLock",
			ExpectedError: commonerrors.ErrLocked,
		},
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				return l.LockWithTimeout(ctx, 200*time.Millisecond)
			},
			Name:          "LockWithTimeout",
			ExpectedError: commonerrors.ErrTimeout,
		},
	}
	test := func(t0 *testing.T, ctx context.Context, fs FS, LockFunc func(l ILock, ctx context.Context) error, expectedError error) {
		dirToLock, err := fs.TempDirInTempDir(fmt.Sprintf("test-lock-dir-%v", faker.DomainName()))
		require.Nil(t0, err)
		defer func() { _ = fs.Rm(dirToLock) }()
		id := "lock"
		Lock1 := fs.NewRemoteLockFile(id, dirToLock)
		defer func() { _ = Lock1.Unlock(ctx) }()
		Lock2 := fs.NewRemoteLockFile(id, dirToLock)
		defer func() { _ = Lock2.Unlock(ctx) }()

		err = Lock1.Unlock(ctx)
		require.Nil(t0, err)
		err = Lock2.Unlock(ctx)
		require.Nil(t0, err)

		err = LockFunc(Lock1, ctx)
		require.Nil(t0, err)

		err = LockFunc(Lock2, ctx)
		require.ErrorIs(t0, err, expectedError)

		err = Lock1.Unlock(ctx)
		require.Nil(t0, err)

		err = LockFunc(Lock2, ctx)
		require.Nil(t0, err)

		err = Lock2.Unlock(ctx)
		require.Nil(t0, err)

		err = fs.Rm(dirToLock)
		require.Nil(t0, err)
	}
	for i := range lockFuncs {
		lockFunc := lockFuncs[i]
		for j := range FileSystemTypes {
			fsType := FileSystemTypes[j]
			t.Run(fmt.Sprintf("%v_for_fs_%v_and_%v", t.Name(), fsType, lockFunc.Name), func(t *testing.T) {
				defer goleak.VerifyNone(t)
				fs := NewFs(fsType)
				ctx := context.Background()
				for c := 0; c < 5; c++ {
					test(t, ctx, fs, lockFunc.LockFunc, lockFunc.ExpectedError)
				}
			})
		}
	}
}

func TestLockConcurrentSafeguard(t *testing.T) {
	lockFuncs := []struct {
		LockFunc      func(l ILock, ctx context.Context) error
		Name          string
		ExpectedError error
	}{
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				return l.TryLock(ctx)
			},
			Name:          "TryLock",
			ExpectedError: commonerrors.ErrLocked,
		},
		{
			LockFunc: func(l ILock, ctx context.Context) error {
				timeout := 200 * time.Millisecond
				return l.LockWithTimeout(ctx, timeout)
			},
			Name:          "LockWithTimeout",
			ExpectedError: commonerrors.ErrTimeout,
		},
	}
	test := func(t0 *testing.T, ctx context.Context, fs FS, LockFunc func(l ILock, ctx context.Context) error, expectedError error) {
		dirToLock, err := fs.TempDirInTempDir(fmt.Sprintf("test-lock-dir-%v", faker.DomainName()))
		require.Nil(t0, err)
		defer func() { _ = fs.Rm(dirToLock) }()
		id := "lock"
		Lock1 := fs.NewRemoteLockFile(id, dirToLock)
		defer func() { _ = Lock1.Unlock(ctx) }()
		Lock2 := fs.NewRemoteLockFile(id, dirToLock)
		defer func() { _ = Lock2.Unlock(ctx) }()

		err = Lock1.Unlock(ctx)
		require.Nil(t0, err)
		err = Lock2.Unlock(ctx)
		require.Nil(t0, err)

		c1 := make(chan error)
		c2 := make(chan error)

		go func(function func(ILock, context.Context) error, l ILock, ctx context.Context) {
			err := function(l, ctx)
			c1 <- err

		}(LockFunc, Lock1, ctx)

		go func(function func(ILock, context.Context) error, l ILock, ctx context.Context) {
			err := function(l, ctx)
			c2 <- err
		}(LockFunc, Lock2, ctx)

		// One will succeed and the other will keep trying till it times out
		err1 := <-c1
		err2 := <-c2
		if fs.GetType() == InMemoryFS {
			// FIXME it was noticed that there could be some race conditions happening in the in-memory file system
			// see https://github.com/spf13/afero/issues/298
			if err1 != nil {
				errortest.RequireError(t0, err1, expectedError)
			}
			if err2 != nil {
				errortest.RequireError(t0, err2, expectedError)
			}
		} else {
			require.NotEqual(t0, err1, err2)
			if err1 == nil {
				errortest.RequireError(t0, err2, expectedError)
			}
			if err2 == nil {
				errortest.RequireError(t0, err1, expectedError)
			}
		}
	}
	for i := range lockFuncs {
		lockFunc := lockFuncs[i]
		for j := range FileSystemTypes {
			fsType := FileSystemTypes[j]
			t.Run(fmt.Sprintf("%v_for_fs_%v_and_%v", t.Name(), fsType, lockFunc.Name), func(t *testing.T) {
				defer goleak.VerifyNone(t)
				fs := NewFs(fsType)
				ctx := context.Background()
				for c := 0; c < 5; c++ {
					test(t, ctx, fs, lockFunc.LockFunc, lockFunc.ExpectedError)
				}
			})
		}
	}
}

func TestLockWithConcurrentAccess(t *testing.T) {
	lockFuncs := []struct {
		LockFunc func(l ILock, ctx context.Context, t *testing.T)
		Name     string
	}{
		{
			LockFunc: func(l ILock, ctx context.Context, t *testing.T) {
				timeout := 2 * time.Second
				err := l.LockWithTimeout(ctx, timeout)
				require.Nil(t, err)
			},
			Name: "LockWithTimeout",
		},
	}
	test := func(t0 *testing.T, ctx context.Context, fs FS, LockFunc func(l ILock, ctx context.Context, t *testing.T)) {
		dirToLock, err := fs.TempDirInTempDir(fmt.Sprintf("test-lock-dir-%v", faker.DomainName()))
		require.Nil(t0, err)
		defer func() { _ = fs.Rm(dirToLock) }()

		Lock1 := fs.NewRemoteLockFile("lock", dirToLock)
		defer func() { _ = Lock1.Unlock(ctx) }()
		Lock2 := fs.NewRemoteLockFile("lock", dirToLock)
		defer func() { _ = Lock2.Unlock(ctx) }()

		err = Lock1.Unlock(ctx)
		require.Nil(t0, err)
		err = Lock2.Unlock(ctx)
		require.Nil(t0, err)

		lockedCount := atomic.NewInt64(0)

		var waitGroup sync.WaitGroup

		LockWithTimeoutTest := func(l ILock, ctx context.Context) {
			LockFunc(l, ctx, t0)

			lockedCount.Inc()

			// Sleep to give the other lock a chance to attempt to lock
			time.Sleep(time.Duration(random.Intn(100)) * time.Millisecond) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec

			// Unlock so other lock can successfully lock
			err = l.Unlock(ctx)
			require.Nil(t0, err)

			waitGroup.Done()
		}

		waitGroup.Add(1)
		go LockWithTimeoutTest(Lock1, ctx)

		waitGroup.Add(1)
		go LockWithTimeoutTest(Lock2, ctx)

		waitGroup.Wait()
		require.Equal(t0, int64(2), lockedCount.Load())
	}
	for i := range lockFuncs {
		lockFunc := lockFuncs[i]
		for j := range FileSystemTypes {
			fsType := FileSystemTypes[j]
			// FIXME it was noticed that there could be some race conditions happening in the in-memory file system when dealing with concurrency
			// see https://github.com/spf13/afero/issues/298
			if fsType != InMemoryFS {
				t.Run(fmt.Sprintf("%v_for_fs_%v_and_%v", t.Name(), fsType, lockFunc.Name), func(t *testing.T) {
					defer goleak.VerifyNone(t)
					fs := NewFs(fsType)
					ctx := context.Background()
					for c := 0; c < 5; c++ {
						test(t, ctx, fs, lockFunc.LockFunc)
					}
				})
			}
		}
	}
}
