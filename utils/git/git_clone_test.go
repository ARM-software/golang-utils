package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestCloneGitBomb(t *testing.T) {
	tests := []struct {
		name                  string
		url                   string
		err                   error
		limits                ILimits
		maxEntriesChannelSize int
	}{
		// FIXME https://kate.io/blog/git-bomb/ is no longer accessible. Uncomment when a new git bomb is available.
		// /*
		// See: https://kate.io/blog/git-bomb/
		// {
		//	name:                  "git bomb small channel saturated",
		//	url:                   "https://github.com/Katee/git-bomb.git",
		//	err:                   fmt.Errorf("%w: entry channel saturated with tree entries", commonerrors.ErrTooLarge),
		//	limits:                NewLimits(1e10, 1e10, 1e10, 10, 1e10, 1e10),
		//	maxEntriesChannelSize: 1000,
		// },
		// {
		//	name:                  "git bomb large channel",
		//	url:                   "https://github.com/Katee/git-bomb.git",
		//	err:                   fmt.Errorf("%w: maximum file count exceeded", commonerrors.ErrTooLarge),
		//	limits:                NewLimits(1e5, 1e6, 1e2, 100, 1e6, 1e10), // max file size: 100KB, max repo size: 1MB, max file count: 1 hundred, max tree depth 10, max entries 1 million, max true size: 10gb
		//	maxEntriesChannelSize: 25000,
		// },
		// {
		//	name:                  "git bomb seg fault",
		//	url:                   "https://github.com/Katee/git-bomb-segfault.git",
		//	err:                   fmt.Errorf("%w: maximum tree depth exceeded", commonerrors.ErrTooLarge),
		//	limits:                NewLimits(1e5, 1e6, 1e4, 4, 1e6, 1e10), // max file size: 100KB, max repo size: 1MB, max file count: 100 thousand, max tree depth 10, max entries 1 million, max true size: 10gb
		//	maxEntriesChannelSize: 25000,
		// },
		{
			name:                  "git bomb large file count",
			url:                   "https://github.com/way2autotesting/DVLA_AutoTest.git",
			err:                   fmt.Errorf("%w: maximum file count exceeded", commonerrors.ErrTooLarge),
			limits:                NewLimits(1e9, 1e9, 10, 4, 1e9, 1e10), // max file size: 100KB, max repo size: 1MB, max file count: 10, max tree depth 10, max entries 1 million, max true size: 10GB
			maxEntriesChannelSize: 25000,
		},
		{
			name:                  "git bomb max true size",
			url:                   "https://github.com/way2autotesting/DVLA_AutoTest.git",
			err:                   fmt.Errorf("%w: maximum true size exceeded", commonerrors.ErrTooLarge),
			limits:                NewLimits(1e9, 1e9, 10, 4, 1e9, 100), // max file size: 100KB, max repo size: 1MB, max file count: 10, max tree depth 10, max entries 1 million, max true size: 100b
			maxEntriesChannelSize: 25000,
		},

		{
			name:                  "known repo that can meet limits",
			url:                   "https://github.com/bulislaw/TrustZone-DevSummit22-Demo",
			err:                   nil,
			limits:                DefaultLimits(), // max file size: 100MB, max repo size: 10gb, max file count: 1 million, max tree depth 12, max entries 100 million, max true size: 1gb
			maxEntriesChannelSize: 25000,
		},
	}
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			MaxEntriesChannelSize = test.maxEntriesChannelSize
			// Cleanup
			err = fs.Rm(destPath)
			require.NoError(t, err)
			empty, err := fs.IsEmpty(destPath)
			require.NoError(t, err)
			require.True(t, empty)
			// Run test
			cloneOptions := GitActionConfig{
				URL: test.url,
			}
			err = CloneWithLimits(context.Background(), destPath, test.limits, &cloneOptions)
			if test.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
				require.Nil(t, err)
			}
		})
	}

}

func TestCloneNormalRepo(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		limits ILimits
	}{
		{
			name:   "with limits",
			url:    "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
			limits: NewLimits(1e8, 1e10, 1e6, 20, 1e6, 1e10), // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
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
			// Cleanup
			err := fs.Rm(destPath)
			require.NoError(t, err)
			empty, err := fs.IsEmpty(destPath)
			require.NoError(t, err)
			require.True(t, empty)
			// Run test
			cloneOptions := GitActionConfig{
				URL:       test.url,
				Reference: "main",
			}
			err = CloneWithLimits(context.Background(), destPath, test.limits, &cloneOptions)
			require.NoError(t, err)
			isEmpty, err := filesystem.IsEmpty(destPath)
			require.NoError(t, err)
			require.False(t, isEmpty)
		})
	}
}

func TestValidationNormalReposErrors(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		err    error
		limits ILimits
	}{
		{
			name:   "too big file",
			err:    fmt.Errorf("%w: maximum individual file size exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1, 1e10, 1e10, 1e10, 1e10, 1e10),
		},
		{
			name:   "too big repo",
			err:    fmt.Errorf("%w: maximum repository size exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1, 1e10, 1e10, 1e10, 1e10),
		},
		{
			name:   "too many files",
			err:    fmt.Errorf("%w: maximum file count exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1, 1e10, 1e10, 1e10),
		},
		{
			name:   "too deep tree",
			err:    fmt.Errorf("%w: maximum tree depth exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1e10, 1, 1e10, 1e10),
		},
		{
			name:   "too many entries",
			err:    fmt.Errorf("%w: maximum entries count exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1e10, 1e10, 1000, 1e10), // entries must be greater than MaxEntriesChannelSize
		},
		{
			name:   "too large true size",
			err:    fmt.Errorf("%w: maximum true size exceeded", commonerrors.ErrTooLarge),
			limits: NewLimits(1e10, 1e10, 1e10, 1e10, 1e10, 1000), // entries must be greater than MaxEntriesChannelSize
		},
	}

	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-test")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

	// Re-run tests but saturate channel during population
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {

			c := NewCloneObject()
			c.repo = repoTest
			err = c.SetupLimits(test.limits)
			require.NoError(t, err)

			err = c.ValidateRepository(context.Background())
			require.ErrorContains(t, err, test.err.Error())
		})
	}

	// Check that small channel gets saturated before initialisation complete
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%s (saturate channel)", test.name), func(t *testing.T) {
			MaxEntriesChannelSize = 100

			c := NewCloneObject()
			c.repo = repoTest
			err = c.SetupLimits(test.limits)
			require.NoError(t, err)

			err = c.ValidateRepository(context.Background())
			require.ErrorContains(t, err, fmt.Errorf("%w: entry channel saturated before initialisation complete", commonerrors.ErrTooLarge).Error())
		})
	}

	// FIXME enable when a git bomb is created
	//// Check channel saturation during run
	// t.Run("channel saturation during run", func(t *testing.T) {
	//	MaxEntriesChannelSize = 1000
	//	err = fs.Rm(destPath)
	//	require.NoError(t, err)
	//
	//	c := NewCloneObject()
	//	c.repo = repoGitBomb
	//	err = c.SetupLimits(DefaultLimits())
	//	require.NoError(t, err)
	//
	//	err = c.ValidateRepository(context.Background())
	//	require.ErrorContains(t, err, fmt.Errorf("%w: entry channel saturated with tree entries", commonerrors.ErrTooLarge).Error())
	// })
}

func TestCloneNonExistentRepo(t *testing.T) {
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
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6, 1e10) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million

	empty, err := fs.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, empty)

	for i := range tests {
		test := tests[i]
		t.Run(test.url, func(t *testing.T) {
			// Cleanup
			err := fs.Rm(destPath)
			require.NoError(t, err)
			empty, err := fs.IsEmpty(destPath)
			require.NoError(t, err)
			require.True(t, empty)
			// Run test
			cloneOptions := GitActionConfig{
				URL: test.url,
			}
			err = CloneWithLimits(context.Background(), destPath, limits, &cloneOptions)
			require.ErrorContains(t, err, test.errorContains)
		})
	}
}

func TestClone(t *testing.T) {
	// Setup
	MaxEntriesChannelSize = 1000
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-test")
	require.NoError(t, err)
	isEmpty, err := filesystem.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, isEmpty)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6, 1e10) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	c := NewCloneObject()

	// Cleanup and make sure cloning git bomb with no checkout doesn't crash
	t.Run("cloning git bomb with no checkout doesn't crash", func(t *testing.T) {
		err = fs.Rm(destPath)
		require.NoError(t, err)
		empty, err := fs.IsEmpty(destPath)
		require.NoError(t, err)
		require.True(t, empty)
		err = c.SetupLimits(limits)
		require.NoError(t, err)
		err = c.Clone(context.Background(), destPath, &GitActionConfig{
			URL:        "https://github.com/Katee/git-bomb.git",
			NoCheckout: true,
		})
		require.NoError(t, err)
	})
}
