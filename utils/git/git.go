package git

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	MaxEntriesChannelSize = 100000000
)

var (
	Parallelisation = 32 // go test cannot handle more than 1 so we use var so we can manually set it to 1 in the tests
)

type Entry struct {
	TreeEntry object.TreeEntry
	TreeDepth int64
}

type RepositoryConfig struct {
	*git.CloneOptions
	MaxTreeDepth      int64
	MaxRepositorySize int64
	MaxFileCount      int64
	MaxFileSize       int64
	MaxEntries        int64
}

type CloneObject struct {
	cfg  RepositoryConfig
	repo *git.Repository
}

func (c *CloneObject) Initialise(path string, cfg RepositoryConfig) (err error) {
	fs := osfs.New(path)
	storer := memory.NewStorage()
	// Make sure we don't check out
	cfg.NoCheckout = true
	c.cfg = cfg
	c.repo, err = git.Clone(storer, fs, cfg.CloneOptions)
	return
}

func (c *CloneObject) ValidateRepository(ctx context.Context) (err error) {
	trees, err := c.repo.TreeObjects()
	if err != nil {
		log.Fatal(err)
	}

	allEntries := make(chan Entry, MaxEntriesChannelSize)

	for {
		tree, err := trees.Next()
		if err != nil {
			break
		}

		for i := range tree.Entries {
			allEntries <- Entry{
				TreeEntry: tree.Entries[i],
				TreeDepth: 0,
			}
		}
	}

	errs, ctx := errgroup.WithContext(ctx)

	totalSize := atomic.NewInt64(0)
	totalFileCount := atomic.NewInt64(0)
	totalEntries := atomic.NewInt64(0)

	var wg sync.WaitGroup
	for p := 0; p < Parallelisation; p++ {
		wg.Add(1)
		errs.Go(func() (err error) {
			for len(allEntries) > 0 {
				err = parallelisation.DetermineContextError(ctx)
				if err != nil {
					return
				}

				entry, ok := <-allEntries
				if !ok {
					return
				}

				totalEntries.Inc()
				if totalEntries.Load() > c.cfg.MaxEntries {
					err = fmt.Errorf("%w: maximum entries count exceeded [%d > %d]", commonerrors.ErrTooLarge, totalEntries.Load(), c.cfg.MaxEntries)
					return
				}

				if entry.TreeDepth > c.cfg.MaxTreeDepth {
					err = fmt.Errorf("%w: maximum tree depth exceeded [%d > %d]", commonerrors.ErrTooLarge, entry.TreeDepth, c.cfg.MaxTreeDepth)
					return
				}

				mode := entry.TreeEntry.Mode
				switch {
				case mode&0o170000 == 0o40000:
					// Tree
					tree, subErr := c.repo.TreeObject(entry.TreeEntry.Hash)
					if subErr != nil {
						err = subErr
						return
					}
					for i := range tree.Entries {
						allEntries <- Entry{
							TreeEntry: tree.Entries[i],
							TreeDepth: entry.TreeDepth + 1,
						}
					}
				case mode&0o170000 == 0o160000:
					// Commit (i.e., submodule)
				case mode&0o170000 == 0o120000:
					// Symlink
				default:
					// Blob
					blob, subErr := c.repo.BlobObject(entry.TreeEntry.Hash)
					if subErr != nil {
						err = subErr
						return
					}

					totalSize.Add(blob.Size)
					if totalSize.Load() > c.cfg.MaxRepositorySize {
						err = fmt.Errorf("%w: maximum repository size exceeded [%d > %d]", commonerrors.ErrTooLarge, totalSize.Load(), c.cfg.MaxRepositorySize)
						return
					}

					if blob.Size > c.cfg.MaxFileSize {
						err = fmt.Errorf("%w: maximum individual file size exceeded [%d > %d]", commonerrors.ErrTooLarge, blob.Size, c.cfg.MaxFileSize)
						return
					}

					totalFileCount.Inc()
					if totalFileCount.Load() > c.cfg.MaxFileCount {
						err = fmt.Errorf("%w: maximum file count exceeded [%d > %d]", commonerrors.ErrTooLarge, totalFileCount.Load(), c.cfg.MaxFileCount)
						return
					}
				}
			}
			wg.Done()
			return nil
		})
	}
	err = errs.Wait()
	if err != nil {
		return
	}
	wg.Wait()
	return
}

func (c *CloneObject) Checkout(opts *git.CheckoutOptions) (err error) {
	worktree, err := c.repo.Worktree()
	if err != nil {
		return
	}
	return worktree.Checkout(opts)
}

func NewCloneObject(cfg RepositoryConfig, path string) (c CloneObject, err error) {
	err = c.Initialise(path, cfg)
	return
}

func CloneWithLimits(ctx context.Context, url, dir string, limits ILimits) (err error) {
	c, err := NewCloneObject(RepositoryConfig{
		MaxTreeDepth:      limits.GetMaxTreeDepth(),
		MaxRepositorySize: limits.GetMaxTotalSize(),
		MaxFileCount:      limits.GetMaxFileCount(),
		MaxFileSize:       limits.GetMaxFileSize(),
		MaxEntries:        limits.GetMaxEntries(),
		CloneOptions: &git.CloneOptions{
			URL: url,
		},
	}, dir)
	if err != nil {
		return
	}

	err = c.ValidateRepository(ctx)
	if err != nil {
		return
	}

	return c.Checkout(&git.CheckoutOptions{Branch: plumbing.ReferenceName("HEAD")})
}
