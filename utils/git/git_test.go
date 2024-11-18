package git

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/units/multiplication"
	"github.com/ARM-software/golang-utils/utils/units/size"
)

// We will populate these with TestMain so they can be reused within the tests
var (
	validBlobHash plumbing.Hash
	validTreeHash plumbing.Hash
	validTag      *plumbing.Reference
	repoTest      *git.Repository
	// FIXME when we have a git bomb at disposal
	// repoGitBomb      *git.Repository
	validTreeEntries []object.TreeEntry
)

func run(m *testing.M) (code int) {
	fs := filesystem.NewFs(filesystem.StandardFS)
	destBase, err := fs.TempDirInTempDir("git-test")
	if err != nil {
		log.Panic(err)
	}
	destPath := filepath.Join(destBase, "blinky")
	defer func() { _ = fs.Rm(destPath) }()
	destGitBomb := filepath.Join(destBase, "bomb")
	defer func() { _ = fs.Rm(destGitBomb) }()

	// Set up a git bomb
	r1, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL: "https://github.com/Open-CMSIS-Pack/csolution-examples.git",
	})
	if err != nil {
		log.Panic(err)
	}

	// FIXME the following git bomb is no longer accessible. Uncomment when we have created ours https://kate.io/blog/making-your-own-exploding-git-repos/
	// Set up a repo
	// r2, err := git.PlainClone(destGitBomb, false, &git.CloneOptions{
	//	URL:        "https://github.com/Katee/git-bomb.git",
	//	NoCheckout: true,
	// })
	// if err != nil {
	//	log.Panic(err)
	// }
	repoTest = r1
	// repoGitBomb = r2

	// Get a valid tree
	trees, err := repoTest.TreeObjects()
	if err != nil {
		// FIXME panic should not be used as it stops the whole test suite from being run
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
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			log.Println("Cannot find a valid blob: ", err)
			return
		}
	}
	validBlobHash = blobHash

	// Create a valid tag
	tag, err := repoTest.CreateTag("testTag", validBlobHash, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "user",
			Email: "example@arm.com",
			When:  time.Now(),
		},
		Message: "test message for tag.",
	})
	if err != nil {
		log.Panic(err)
	}
	validTag = tag

	// Run the tests
	code = m.Run()
	return
}

// Clone the repository once and use it and the valid trees/blobs/entries in all the tests
func TestMain(m *testing.M) {
	code := run(m) // extract into function so that defers are called before os.Exit
	os.Exit(code)
}

func TestHandleTreeEntry(t *testing.T) {
	// Setup
	MaxEntriesChannelSize = 100000
	c := NewCloneObject()
	limits := NewLimits(100*size.MB, 10*size.GB, multiplication.Mega, 20, multiplication.Mega, 10*size.GB)
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
	limits := NewLimits(100*size.MB, 10*size.GB, multiplication.Mega, 20, multiplication.Mega, 10*size.GB)
	err := c.SetupLimits(limits)
	require.NoError(t, err)
	c.repo = repoTest
	totalSize := atomic.NewInt64(0)
	totalFileCount := atomic.NewInt64(0)
	totalTrueSize := atomic.NewInt64(0)

	// Test normal
	t.Run("normal", func(t *testing.T) {
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount, totalTrueSize)
		require.NoError(t, err)
		require.Equal(t, int64(1), totalFileCount.Load())
		require.True(t, totalSize.Load() > 0)
	})

	// Test whether too large blob returns error
	t.Run("too large blob returns error", func(t *testing.T) {
		limits = NewLimits(0, 10*size.GB, multiplication.Mega, 20, multiplication.Mega, 10*size.GB)
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		c.seen = mapset.NewSet[plumbing.Hash]()
		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		totalTrueSize = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount, totalTrueSize)
		require.Error(t, err)
		assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	})

	// Test whether too many files returns error
	t.Run("too many files returns error", func(t *testing.T) {
		limits = NewLimits(100*size.KB, 10*size.GB, 0, 20, multiplication.Mega, 10*size.GB)
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		c.seen = mapset.NewSet[plumbing.Hash]()
		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		totalTrueSize = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount, totalTrueSize)
		require.Error(t, err)
		assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	})

	// Test whether too large repo fails
	t.Run("too large repo fails", func(t *testing.T) {
		limits = NewLimits(100*size.KB, 0, multiplication.Mega, 20, multiplication.Mega, 10*size.GB)
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		c.seen = mapset.NewSet[plumbing.Hash]()
		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		totalTrueSize = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount, totalTrueSize)
		require.Error(t, err)
		assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	})

	// Test whether too large repo fails based on true size
	t.Run("too large repo fails based on true size", func(t *testing.T) {
		limits = NewLimits(100*size.KB, 10*size.GB, multiplication.Mega, 20, multiplication.Mega, 0)
		err = c.SetupLimits(limits)
		require.NoError(t, err)

		c.seen = mapset.NewSet[plumbing.Hash]()
		totalSize = atomic.NewInt64(0)
		totalFileCount = atomic.NewInt64(0)
		totalTrueSize = atomic.NewInt64(0)
		err = c.handleBlobEntry(Entry{
			TreeEntry: object.TreeEntry{
				Name: "test",
				Hash: validBlobHash,
				Mode: 0,
			},
			TreeDepth: 0,
		}, totalSize, totalFileCount, totalTrueSize)
		require.Error(t, err)
		assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	})
}

func TestCheckDepthAndTotalEntries(t *testing.T) {
	// Setup
	MaxEntriesChannelSize = 100000
	c := NewCloneObject()
	limits := NewLimits(100*size.MB, 10*size.GB, multiplication.Mega, 10, multiplication.Mega, 10*size.GB)
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
		assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	})

	// Check too many entries
	t.Run("too many entries", func(t *testing.T) {
		limits = NewLimits(100*size.MB, 10*size.GB, multiplication.Mega, 20, 0, 10*size.GB)
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
		assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	})
}

func TestPopulateInitialEntries(t *testing.T) {
	// Setup
	c := NewCloneObject()
	limits := NewLimits(100*size.MB, 10*size.GB, multiplication.Mega, 20, multiplication.Mega, 10*size.GB)
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

	// FIXME uncomment when set for the repository
	// // Check unsuccessful population sue to channel size
	// t.Run("unsuccessful population sue to channel size", func(t *testing.T) {
	//	MaxEntriesChannelSize = 100
	//	c = NewCloneObject()
	//	err = c.SetupLimits(NewLimits(1e8, 1e10, 1e6, 20, 1e6, 1e10)) // max file size: 100MB, max repo size: 10GB, max file count: 1 million, max tree depth 1, max entries 1 million, max true size 10GB
	//	require.NoError(t, err)
	//	c.repo = repoTest
	//	require.Empty(t, c.allEntries)
	//	err = c.populateInitialEntries(context.Background())
	//	require.Error(t, err)
	//	assert.True(t, commonerrors.Any(err, commonerrors.ErrTooLarge))
	// })
}

func TestParseReference(t *testing.T) {
	tests := []struct {
		reference string
		branch    plumbing.ReferenceName
		hash      plumbing.Hash
	}{
		{
			reference: "main",
			branch:    plumbing.NewBranchReferenceName("main"),
			hash:      plumbing.NewHash(""),
		},
		{
			reference: "refs/heads/main",
			branch:    plumbing.NewBranchReferenceName("main"),
			hash:      plumbing.NewHash(""),
		},
		{
			reference: "ref: refs/heads/main",
			branch:    plumbing.NewBranchReferenceName("main"),
			hash:      plumbing.NewHash(""),
		},
		{
			reference: validBlobHash.String(),
			branch:    plumbing.ReferenceName(""),
			hash:      validBlobHash,
		},
		{
			reference: validTag.Name().String(),
			branch:    validTag.Name(),
			hash:      plumbing.NewHash(""),
		},
		{
			reference: "",
			branch:    plumbing.HEAD,
			hash:      plumbing.NewHash(""),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%v_parseReference(%v)", i, test.reference), func(t *testing.T) {
			c := NewCloneObject()
			err := c.SetupLimits(DefaultLimits())
			require.NoError(t, err)
			c.repo = repoTest
			branch, hash := c.parseReference(&GitActionConfig{
				Reference: test.reference,
			})
			require.Equal(t, test.branch, branch)
			require.Equal(t, test.hash, hash)
		})
	}
}
