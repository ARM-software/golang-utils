package git

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// We will populate these with TestMain so they can be reused within the tests
var (
	validBlobHash    plumbing.Hash
	validTreeHash    plumbing.Hash
	repoTest         *git.Repository
	validTreeEntries []object.TreeEntry
)

// Clone the repository once and use it and the valid trees/blobs/entries in all the tests
func TestMain(m *testing.M) {
	fs := filesystem.NewFs(filesystem.StandardFS)
	destPath, err := fs.TempDirInTempDir("git-bomb")
	if err != nil {
		log.Panic(err)
	}
	defer func() { _ = fs.Rm(destPath) }()

	// Set up a repo
	r, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: "https://github.com/Arm-Examples/Blinky_MIMXRT1064-EVK_RTX",
	})
	if err != nil {
		log.Panic(err)
	}
	repoTest = r

	// Get a valid tree
	trees, err := repoTest.TreeObjects()
	if err != nil {
		log.Panic(err)
	}
	tree, err := trees.Next()
	if err != nil {
		log.Panic(err)
	}
	if tree == nil {
		log.Panic(fmt.Errorf("%w: tree undefined", commonerrors.ErrUndefined))
	}
	if len(tree.Entries) == 0 {
		log.Panic(fmt.Errorf("%w: tree entries empty", commonerrors.ErrEmpty))
	}

	validTreeHash = tree.Hash
	validTreeEntries = tree.Entries

	// Get a valid blob
	blobHash, err := getValidBlobHash(tree)
	if err != nil {
		log.Panic(err)
	}
	validBlobHash = blobHash

	// Run the tests
	_ = m.Run()
}

func TestHandleTreeEntry(t *testing.T) {
	// Setup
	MaxEntriesChannelSize = 100000
	c := NewCloneObject()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	err := c.SetupLimits(limits)
	require.NoError(t, err)
	c.repo = repoTest

	// Check entries are added
	oldLen := len(c.allEntries)
	err = c.handleTreeEntry(Entry{
		TreeEntry: object.TreeEntry{
			Name: "test",
			Hash: validTreeHash,
			Mode: 0,
		},
		TreeDepth: 0,
	})
	require.NoError(t, err)
	newLen := len(c.allEntries)
	require.Equal(t, oldLen+len(validTreeEntries), newLen)
}

func getValidBlobHash(tree *object.Tree) (blobHash plumbing.Hash, err error) {
	for i := range tree.Entries {
		entry := tree.Entries[i]
		if entry.Mode.IsFile() {
			blobHash = entry.Hash
			return
		}
	}
	err = commonerrors.ErrNotFound
	return
}

func TestHandleBlobEntry(t *testing.T) {
	// Setup
	MaxEntriesChannelSize = 100000
	c := NewCloneObject()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	err := c.SetupLimits(limits)
	require.NoError(t, err)
	c.repo = repoTest
	totalSize := atomic.NewInt64(0)
	totalFileCount := atomic.NewInt64(0)

	// Test normal
	t.Run("normal", func(t *testing.T) {
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount)
		require.NoError(t, err)
		require.Equal(t, int64(1), totalFileCount.Load())
		require.True(t, totalSize.Load() > 0)
	})

	// Test whether too large blob returns error
	t.Run("too large blob returns error", func(t *testing.T) {
		limits = NewLimits(0, 1e10, 1e6, 20, 1e6) // max file size: 0, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Errorf("%w: maximum individual file size exceeded", commonerrors.ErrTooLarge).Error())
	})

	// Test whether too many files returns error
	t.Run("too many files returns error", func(t *testing.T) {
		limits = NewLimits(1e5, 1e10, 0, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 0, max tree depth 1, max entries 1 million
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Errorf("%w: maximum file count exceeded", commonerrors.ErrTooLarge).Error())
	})

	// Test whether too large repo fails
	t.Run("too large repo fails", func(t *testing.T) {
		limits = NewLimits(1e5, 0, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 0, max file count: 1 million, max tree depth 1, max entries 1 million
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Errorf("%w: maximum repository size exceeded", commonerrors.ErrTooLarge).Error())
	})
}

func TestCheckDepthAndTotalEntries(t *testing.T) {
	// Setup
	MaxEntriesChannelSize = 100000
	c := NewCloneObject()
	limits := NewLimits(1e8, 1e10, 1e6, 10, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	err := c.SetupLimits(limits)
	require.NoError(t, err)
	c.repo = repoTest
	totalEntries := atomic.NewInt64(0)

	// Check normal
	t.Run("normal", func(t *testing.T) {
		err = c.checkDepthAndTotalEntries(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalEntries)
		require.NoError(t, err)
		require.Equal(t, int64(1), totalEntries.Load())
		require.True(t, totalEntries.Load() > 0)
	})

	// Check too deep depth
	t.Run("too deep depth", func(t *testing.T) {
		totalEntries = atomic.NewInt64(0)
		err = c.checkDepthAndTotalEntries(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: limits.GetMaxTreeDepth() + 1,
		}, totalEntries)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Errorf("%w: maximum tree depth exceeded", commonerrors.ErrTooLarge).Error())
	})

	// Check too many entries
	t.Run("too many entries", func(t *testing.T) {
		limits = NewLimits(1e8, 1e10, 1e6, 20, 0) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 0
		err = c.SetupLimits(limits)
		require.NoError(t, err)
		totalEntries = atomic.NewInt64(0)
		err = c.checkDepthAndTotalEntries(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalEntries)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Errorf("%w: maximum entries count exceeded", commonerrors.ErrTooLarge).Error())
	})
}

func TestPopulateInitialEntries(t *testing.T) {
	// Setup
	c := NewCloneObject()
	limits := NewLimits(1e8, 1e10, 1e6, 20, 1e6) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
	err := c.SetupLimits(limits)
	require.NoError(t, err)
	c.repo = repoTest

	// make sure tree has content
	trees, err := c.repo.TreeObjects()
	require.NoError(t, err)
	tree, err := trees.Next()
	require.NoError(t, err)
	require.NotNil(t, tree)
	require.NotEmpty(t, tree.Entries)
	require.NoError(t, err)

	// Check successful population
	t.Run("successful population", func(t *testing.T) {
		MaxEntriesChannelSize = 10000
		require.Empty(t, c.allEntries)
		err = c.populateInitialEntries(context.Background())
		require.NoError(t, err)
		require.True(t, len(c.allEntries) > 0)
	})

	// Check unsuccessful population sue to channel size
	t.Run("unsuccessful population sue to channel size", func(t *testing.T) {
		MaxEntriesChannelSize = 100
		c = NewCloneObject()
		err = c.SetupLimits(NewLimits(1e8, 1e10, 1e6, 20, 1e6)) // max file size: 100MB, max repo size: 1GB, max file count: 1 million, max tree depth 1, max entries 1 million
		c.repo = repoTest
		require.NoError(t, err)
		require.Empty(t, c.allEntries)
		err = c.populateInitialEntries(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Errorf("%w: entry channel saturated before initialisation complete", commonerrors.ErrTooLarge).Error())
	})
}
