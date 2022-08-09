package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestHandleTreeEntry(t *testing.T) {
	// Setup
	var c CloneObject
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	repo, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
	})
	require.NoError(t, err)
	c.repo = repo

	// Get a valid tree
	trees, err := c.repo.TreeObjects()
	require.NoError(t, err)
	tree, err := trees.Next()
	require.NoError(t, err)
	require.NotNil(t, tree)
	require.NotEmpty(t, tree.Entries)
	require.NoError(t, err)

	// Check entries are added
	oldLen := len(c.allEntries)
	err = c.handleTreeEntry(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: tree.Hash,
			Mode: 0,
		},
		TreeDepth: 0,
	})
	require.NoError(t, err)
	newLen := len(c.allEntries)
	require.Equal(t, oldLen+len(tree.Entries), newLen)
}

func getValidBlobHash(tree *object.Tree) (blobHash plumbing.Hash, err error) {
	for i := range tree.Entries {
		entry := tree.Entries[i]
		if entry.Mode.IsFile() {
			fmt.Println(entry)
			blobHash = entry.Hash
			return
		}
	}
	err = commonerrors.ErrNotFound
	return
}

func TestHandleBlobEntry(t *testing.T) {
	// Setup
	var c CloneObject
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	repo, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
	})
	require.NoError(t, err)
	c.repo = repo

	// Get a valid tree
	trees, err := c.repo.TreeObjects()
	require.NoError(t, err)
	tree, err := trees.Next()
	require.NoError(t, err)
	require.NotNil(t, tree)
	require.NotEmpty(t, tree.Entries)
	require.NoError(t, err)

	// Get a valid blob hash
	blobHash, err := getValidBlobHash(tree)
	require.NoError(t, err)
	fmt.Println(blobHash)

	// Test normal
	totalSize := atomic.NewInt64(0)
	totalFileCount := atomic.NewInt64(0)
	err = c.handleBlobEntry(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: 0,
	}, totalSize, totalFileCount)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalFileCount.Load())
	require.True(t, totalSize.Load() > 0)

	// Test whether too large blob returns error
	limits = NewLimits(0, 1e10, 1e6, 20, 1e6) // max file size: 0, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})

	totalSize = atomic.NewInt64(0)
	totalFileCount = atomic.NewInt64(0)
	err = c.handleBlobEntry(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: 0,
	}, totalSize, totalFileCount)
	require.Error(t, err)
	require.ErrorContains(t, err, fmt.Errorf("%w: maximum individual file size exceeded", commonerrors.ErrTooLarge).Error())

	// Test whether too many files returns error
	limits = NewLimits(1e5, 1e10, 0, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 0, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})

	totalSize = atomic.NewInt64(0)
	totalFileCount = atomic.NewInt64(0)
	err = c.handleBlobEntry(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: 0,
	}, totalSize, totalFileCount)
	require.Error(t, err)
	require.ErrorContains(t, err, fmt.Errorf("%w: maximum file count exceeded", commonerrors.ErrTooLarge).Error())

	// Test whether too large repo fails
	limits = NewLimits(1e5, 0, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 0, max file count: 1 million, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})

	totalSize = atomic.NewInt64(0)
	totalFileCount = atomic.NewInt64(0)
	err = c.handleBlobEntry(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: 0,
	}, totalSize, totalFileCount)
	require.Error(t, err)
	require.ErrorContains(t, err, fmt.Errorf("%w: maximum repository size exceeded", commonerrors.ErrTooLarge).Error())
}

func TestCheckDepthAndTotalEntries(t *testing.T) {
	// Setup
	var c CloneObject
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	repo, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
	})
	require.NoError(t, err)
	c.repo = repo

	// Get a valid tree
	trees, err := c.repo.TreeObjects()
	require.NoError(t, err)
	tree, err := trees.Next()
	require.NoError(t, err)
	require.NotNil(t, tree)
	require.NotEmpty(t, tree.Entries)
	require.NoError(t, err)

	// Get a valid blob hash
	blobHash, err := getValidBlobHash(tree)
	require.NoError(t, err)

	// Check normal
	totalEntries := atomic.NewInt64(0)
	err = c.checkDepthAndTotalEntries(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: 0,
	}, totalEntries)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalEntries.Load())
	require.True(t, totalEntries.Load() > 0)

	// Check too deep depth
	totalEntries = atomic.NewInt64(0)
	err = c.checkDepthAndTotalEntries(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: limits.GetMaxTreeDepth() + 1,
	}, totalEntries)
	require.Error(t, err)
	require.ErrorContains(t, err, fmt.Errorf("%w: maximum tree depth exceeded", commonerrors.ErrTooLarge).Error())

	// Check too many entries
	limits = NewLimits(1e8, 1e10, 1e6, 20, 0) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 0
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	totalEntries = atomic.NewInt64(0)
	err = c.checkDepthAndTotalEntries(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: blobHash,
			Mode: 0,
		},
		TreeDepth: 0,
	}, totalEntries)
	require.Error(t, err)
	require.ErrorContains(t, err, fmt.Errorf("%w: maximum entries count exceeded", commonerrors.ErrTooLarge).Error())
}

func TestPopulateInitialEntries(t *testing.T) {
	// Setup
	var c CloneObject
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	repo, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
	})
	require.NoError(t, err)
	c.repo = repo

	// make sure tree has content
	trees, err := c.repo.TreeObjects()
	require.NoError(t, err)
	tree, err := trees.Next()
	require.NoError(t, err)
	require.NotNil(t, tree)
	require.NotEmpty(t, tree.Entries)
	require.NoError(t, err)

	// Check successful population
	err = c.populateInitialEntries(context.Background())
	require.NoError(t, err)
	require.True(t, len(c.allEntries) > 0)
}

func TestClone(t *testing.T) {
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-test")
	require.NoError(t, err)
	isEmpty, err := filesystem.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, isEmpty)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	branch := "main"
	var c CloneObject
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	err = c.Clone(context.Background(), destPath, &GitActionConfig{
		URL:    "https://github.com/Katee/git-bomb.git",
		Branch: "main",
	})
	require.NoError(t, err)
	isEmpty, err = filesystem.IsEmpty(destPath)
	require.NoError(t, err)
	require.False(t, isEmpty)
	head, err := c.repo.Head()
	require.NoError(t, err)
	require.Equal(t, plumbing.NewBranchReferenceName(branch), head.Name())

	// Cleanup and make sure cloning git bomb with no checkout doesn't crash
	err = fs.Rm(destPath)
	require.NoError(t, err)
	empty, err := fs.IsEmpty(destPath)
	require.NoError(t, err)
	require.True(t, empty)
	c.SetupLimits(RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	err = c.Clone(context.Background(), destPath, &GitActionConfig{
		URL:        "https://github.com/Katee/git-bomb.git",
		NoCheckout: true,
	})
	require.NoError(t, err)
}

func TestCloneGitBomb(t *testing.T) {
	ValidationParallelisation = 1 // so go test doesn't break
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
	limits := NewLimits(1e5, 1e6, 1e5, 10, 1e5) // max file size: 100KB, max repo size: 1MB, max file count: 100 thousand, max tree depth 10, max entries 100 thousand

	for i := range tests {
		test := tests[i]
		t.Run(test.url, func(t *testing.T) {
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
			err = CloneWithLimits(context.Background(), destPath, limits, &cloneOptions)
			require.Error(t, err)
			require.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
		})
	}
}

func TestCloneNormalRepo(t *testing.T) {
	ValidationParallelisation = 1 // so go test doesn't break
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
			// Cleanup
			err := fs.Rm(destPath)
			require.NoError(t, err)
			empty, err := fs.IsEmpty(destPath)
			require.NoError(t, err)
			require.True(t, empty)
			// Run test
			cloneOptions := GitActionConfig{
				URL:    test.url,
				Branch: "main",
			}
			err = CloneWithLimits(context.Background(), destPath, test.limits, &cloneOptions)
			require.NoError(t, err)
			isEmpty, err := filesystem.IsEmpty(destPath)
			require.NoError(t, err)
			require.False(t, isEmpty)
		})
	}
}

func TestCloneNormalReposErrors(t *testing.T) {
	ValidationParallelisation = 1 // so go test doesn't break
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

	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-test")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

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
				URL:    test.url,
				Branch: "main",
			}
			err = CloneWithLimits(context.Background(), destPath, test.limits, &cloneOptions)
			require.ErrorContains(t, err, test.err.Error())
		})
	}
}

func TestCloneNonExistentRepo(t *testing.T) {
	ValidationParallelisation = 1 // so go test doesn't break
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
