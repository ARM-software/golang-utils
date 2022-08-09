package git

import (
	"context"
	"fmt"
	"io"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	MaxEntriesChannelSize = 100000000
)

var (
	ValidationParallelisation = 32 // go test cannot handle more than 1 so we use var so we can manually set it to 1 in the tests
)

type Entry struct {
	TreeEntry object.TreeEntry
	TreeDepth int64
}

type RepositoryLimitsConfig struct {
	maxTreeDepth      int64
	maxRepositorySize int64
	maxFileCount      int64
	maxFileSize       int64
	maxEntries        int64
}

type CloneObject struct {
	cfg        *RepositoryLimitsConfig
	repo       *git.Repository
	allEntries chan Entry
}

func (c *CloneObject) handleTreeEntry(entry Entry) (err error) {
	tree, subErr := c.repo.TreeObject(entry.TreeEntry.Hash)
	if subErr != nil {
		err = subErr
		return
	}
	for i := range tree.Entries {
		c.allEntries <- Entry{
			TreeEntry: tree.Entries[i],
			TreeDepth: entry.TreeDepth + 1,
		}
	}
	return
}

func (c *CloneObject) handleCommitEntry(entry Entry) (err error) {
	// Unknown if necessary. Add code here if necessary in future
	return
}

func (c *CloneObject) handleSymlinkEntry(entry Entry) (err error) {
	// Unknown if necessary. Add code here if necessary in future
	return
}

func (c *CloneObject) handleBlobEntry(entry Entry, totalSize *atomic.Int64, totalFileCount *atomic.Int64) (err error) {
	blob, subErr := c.repo.BlobObject(entry.TreeEntry.Hash)
	if subErr != nil {
		err = subErr
		return
	}

	totalSize.Add(blob.Size)
	if totalSize.Load() > c.cfg.maxRepositorySize {
		err = fmt.Errorf("%w: maximum repository size exceeded [%d > %d]", commonerrors.ErrTooLarge, totalSize.Load(), c.cfg.maxRepositorySize)
		return
	}

	if blob.Size > c.cfg.maxFileSize {
		err = fmt.Errorf("%w: maximum individual file size exceeded [%d > %d]", commonerrors.ErrTooLarge, blob.Size, c.cfg.maxFileSize)
		return
	}

	totalFileCount.Inc()
	if totalFileCount.Load() > c.cfg.maxFileCount {
		err = fmt.Errorf("%w: maximum file count exceeded [%d > %d]", commonerrors.ErrTooLarge, totalFileCount.Load(), c.cfg.maxFileCount)
		return
	}
	return
}

func (c *CloneObject) checkDepthAndTotalEntries(entry Entry, totalEntries *atomic.Int64) (err error) {
	totalEntries.Inc()
	if totalEntries.Load() > c.cfg.maxEntries {
		err = fmt.Errorf("%w: maximum entries count exceeded [%d > %d]", commonerrors.ErrTooLarge, totalEntries.Load(), c.cfg.maxEntries)
		return
	}

	if entry.TreeDepth > c.cfg.maxTreeDepth {
		err = fmt.Errorf("%w: maximum tree depth exceeded [%d > %d]", commonerrors.ErrTooLarge, entry.TreeDepth, c.cfg.maxTreeDepth)
		return
	}
	return
}

func (c *CloneObject) populateInitialEntries(ctx context.Context) (err error) {
	trees, err := c.repo.TreeObjects()
	if err != nil {
		return
	}

	for {
		err = parallelisation.DetermineContextError(ctx)
		if err != nil {
			return
		}

		tree, subErr := trees.Next()
		if subErr != nil {
			if commonerrors.Any(subErr, io.EOF) {
				break
			} else {
				err = subErr
				return
			}
		}

		for i := range tree.Entries {
			c.allEntries <- Entry{
				TreeEntry: tree.Entries[i],
				TreeDepth: 0,
			}
		}
	}
	return
}

func (c *CloneObject) SetupLimits(cfg *RepositoryLimitsConfig) (err error) {
	if cfg == nil {
		return fmt.Errorf("%w: limits config undefined", commonerrors.ErrUndefined)
	}
	c.cfg = cfg
	c.allEntries = make(chan Entry, MaxEntriesChannelSize)
	return
}

func NewCloneObject() *CloneObject {
	limits := NoLimits()
	return &CloneObject{
		cfg: &RepositoryLimitsConfig{
			maxTreeDepth:      limits.GetMaxTreeDepth(),
			maxRepositorySize: limits.GetMaxTotalSize(),
			maxFileCount:      limits.GetMaxFileCount(),
			maxFileSize:       limits.GetMaxFileSize(),
			maxEntries:        limits.GetMaxEntries(),
		},
	}
}

// Clone without checkout or validation
func (c *CloneObject) Clone(ctx context.Context, path string, cfg *GitActionConfig) (err error) {
	recursiveSubModules := git.NoRecurseSubmodules
	if cfg.GetRecursiveSubModules() {
		recursiveSubModules = git.DefaultSubmoduleRecursionDepth
	}

	c.repo, err = git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		NoCheckout:        cfg.GetNoCheckout(),
		URL:               cfg.GetURL(),
		Auth:              cfg.GetAuth(),
		RemoteName:        "",
		ReferenceName:     "",
		SingleBranch:      false,
		Depth:             cfg.GetDepth(),
		RecurseSubmodules: recursiveSubModules,
		Progress:          nil,
		Tags:              cfg.GetTags(),
		InsecureSkipTLS:   false,
		CABundle:          nil,
	})
	return
}

// After cloning without checkout, valdiate the repository to check for git bombs
func (c *CloneObject) ValidateRepository(ctx context.Context) (err error) {
	if err = c.populateInitialEntries(ctx); err != nil {
		return
	}

	errs, ctx := errgroup.WithContext(ctx)

	totalSize := atomic.NewInt64(0)
	totalFileCount := atomic.NewInt64(0)
	totalEntries := atomic.NewInt64(0)

	for p := 0; p < ValidationParallelisation; p++ {
		errs.Go(func() (err error) {
			for len(c.allEntries) > 0 {
				err = parallelisation.DetermineContextError(ctx)
				if err != nil {
					return
				}

				entry, ok := <-c.allEntries
				if !ok {
					return
				}

				if err = c.checkDepthAndTotalEntries(entry, totalEntries); err != nil {
					return
				}

				mode := entry.TreeEntry.Mode
				switch {
				case mode&0o170000 == 0o40000:
					if err = c.handleTreeEntry(entry); err != nil {
						return
					}
				case mode&0o170000 == 0o160000:
					// Commit (i.e., submodule)
					if err = c.handleCommitEntry(entry); err != nil {
						return
					}
				case mode&0o170000 == 0o120000:
					// Symlink
					if err = c.handleSymlinkEntry(entry); err != nil {
						return
					}
				default:
					// Blob
					if err = c.handleBlobEntry(entry, totalSize, totalFileCount); err != nil {
						return
					}
				}
			}
			return nil
		})
	}
	err = errs.Wait()
	return
}

func (c *CloneObject) Checkout(gitOptions *GitActionConfig) (err error) {
	worktree, err := c.repo.Worktree()
	if err != nil {
		return
	}
	var branch plumbing.ReferenceName
	if gitOptions.GetBranch() != "" {
		branch = plumbing.NewBranchReferenceName(gitOptions.Branch)
	}
	var hash plumbing.Hash
	if gitOptions.GetHash() != "" {
		hash = plumbing.NewHash(gitOptions.Hash)
	}
	checkoutOptions := git.CheckoutOptions{
		Hash:   hash,
		Branch: branch,
		Create: gitOptions.GetCreate(),
	}
	return worktree.Checkout(&checkoutOptions)
}

// Clone a repository with limits on the max tree depth, the max repository size, the max file count, the max individual file size, and the max entries
func CloneWithLimits(ctx context.Context, dir string, limits ILimits, gitOptions *GitActionConfig) (err error) {
	c := NewCloneObject()
	err = c.SetupLimits(&RepositoryLimitsConfig{
		maxTreeDepth:      limits.GetMaxTreeDepth(),
		maxRepositorySize: limits.GetMaxTotalSize(),
		maxFileCount:      limits.GetMaxFileCount(),
		maxFileSize:       limits.GetMaxFileSize(),
		maxEntries:        limits.GetMaxEntries(),
	})
	if err != nil {
		return
	}

	gitOptions.NoCheckout = true // don't checkout so we can validate it
	err = c.Clone(ctx, dir, gitOptions)
	if err != nil {
		return
	}

	err = c.ValidateRepository(ctx)
	if err != nil {
		return
	}

	return c.Checkout(gitOptions)
}
