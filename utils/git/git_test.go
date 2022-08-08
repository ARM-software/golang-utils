package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/require"
)

func TestCloneGitBomb(t *testing.T) {
	Parallelisation = 1 // so go test doesn't break
	tests := []struct {
		url string
	}{
		// See: https://kate.io/blog/git-bomb/
		{
			url: "https://github.com/Katee/git-bomb.git",
		},
		{
			url: "https://github.com/Katee/git-bomb-segfault.git",
		},
	}
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e5, 1e6, 1e6, 20, 1e6) // max file size: 100KB, max repo size: 1MB, max file count: 1 million, max tree depth 1, max entries 1 million

	empty, err := fs.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, empty)

	for i := range tests {
		test := tests[i]
		t.Run(test.url, func(t *testing.T) {
			err := CloneWithLimits(context.Background(), test.url, destPath, limits)
			require.Error(t, err)
			require.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
		})
	}

	// check not cloned
	empty, err = fs.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, empty)
}

func TestCloneNormalRepo(t *testing.T) {
	Parallelisation = 1 // so go test doesn't break
	tests := []struct {
		name   string
		url    string
		limits ILimits
	}{
		{
			name:   "with limits",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			limits: NewLimits(1e8, 1e10, 1e6, 20, 1e6), // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
		},
		{
			name:   "no limits",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			limits: NoLimits(),
		},
	}
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-test")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

	empty, err := fs.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, empty)

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			err := CloneWithLimits(context.Background(), test.url, destPath, test.limits)
			require.NoError(t, err)
			isEmpty, err := filesystem.IsEmpty(destPath)
			require.NoError(t, err)
			require.False(t, isEmpty)
		})
	}
}

func TestCloneNormalReposErrors(t *testing.T) {
	Parallelisation = 1 // so go test doesn't break
	tests := []struct {
		name   string
		url    string
		err    error
		limits ILimits
	}{
		{
			name:   "too big file",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			err:    fmt.Errorf("%w: maximum individual file size exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1, 1e10, 1e10, 1e10, 1e10),
		},
		{
			name:   "too big repo",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			err:    fmt.Errorf("%w: maximum repository size exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1, 1e10, 1e10, 1e10),
		},
		{
			name:   "too many files",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			err:    fmt.Errorf("%w: maximum file count exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1, 1e10, 1e10),
		},
		{
			name:   "too deep tree",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			err:    fmt.Errorf("%w: maximum tree depth exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1e10, 1, 1e10),
		},
		{
			name:   "too many entries",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			err:    fmt.Errorf("%w: maximum entries count exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1e10, 1e10, 1),
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			err := CloneWithLimits(context.Background(), test.url, "test", test.limits)
			require.ErrorContains(t, err, test.err.Error())
		})
	}
}

func TestCloneNonExistentRepo(t *testing.T) {
	Parallelisation = 1 // so go test doesn't break
	tests := []struct {
		url           string
		errorContains string
	}{
		{
			url:           "https://github.com/Arm-Examples/Fake-Repo",
			errorContains: "authentication required", // Because GitHub will assume it is a private repository
		},
		{
			url:           "fake.url.com/fake-repo",
			errorContains: "repository not found",
		},
	}
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-test")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million

	empty, err := fs.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, empty)

	for i := range tests {
		test := tests[i]
		t.Run(test.url, func(t *testing.T) {
			err := CloneWithLimits(context.Background(), test.url, destPath, limits)
			require.ErrorContains(t, err, test.errorContains)
		})
	}
}
