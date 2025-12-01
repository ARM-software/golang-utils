package git

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	refPrefix    = "refs/"
	symrefPrefix = "ref: "
)

var (
	// Variables so it can be modified in testing

	MaxEntriesChannelSize     = 100000
	ValidationParallelisation = 32
)

type Entry struct {
	TreeEntry object.TreeEntry
	TreeDepth int64
	Seen      string
}

type CloneObject struct {
	cfg        ILimits
	mu         sync.Mutex // (*git.Repository).XXXObject(h plumbing.Hash) are not thread safe
	repo       *git.Repository
	allEntries chan Entry

	seen mapset.Set[plumbing.Hash]

	totalSize      *atomic.Int64
	trueSize       *atomic.Int64
	totalFileCount *atomic.Int64
	totalEntries   *atomic.Int64

	nonTreeOnlyMutex   chan int // TODO: when updated to go1.18 use sync.Mutex
	processNonTreeOnly *atomic.Bool
	treeSeenIdentifier *atomic.String
}

func (c *CloneObject) getTreeObject(hash plumbing.Hash) (tree *object.Tree, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	tree, err = c.repo.TreeObject(hash)
	return
}

func (c *CloneObject) getBlobObject(hash plumbing.Hash) (blob *object.Blob, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	blob, err = c.repo.BlobObject(hash)
	return
}

func (c *CloneObject) handleEntry(entry Entry) (err error) {
	mode := entry.TreeEntry.Mode
	switch mode & 0o170000 {
	case 0o40000:
		if c.processNonTreeOnly.Load() {
			seenIdentifier := c.treeSeenIdentifier.Load()
			if entry.Seen == seenIdentifier {
				err = fmt.Errorf("%w: entry channel saturated with tree entries", commonerrors.ErrTooLarge)
				return
			}
			entry.Seen = seenIdentifier
			c.allEntries <- entry
			return
		}
		if err = c.handleTreeEntry(entry); err != nil {
			return
		}
	case 0o160000:
		// Commit (i.e., submodule)
		if err = c.handleCommitEntry(entry); err != nil {
			return
		}
	case 0o120000:
		// Symlink
		if err = c.handleSymlinkEntry(entry); err != nil {
			return
		}
	default:
		// Blob
		if err = c.handleBlobEntry(entry, c.totalSize, c.totalFileCount, c.trueSize); err != nil {
			return
		}
	}
	return
}

func (c *CloneObject) handleTreeEntry(entry Entry) (err error) {
	tree, subErr := c.getTreeObject(entry.TreeEntry.Hash)
	if subErr != nil {
		err = subErr
		return
	}
	for i := range tree.Entries {
		c.allEntries <- Entry{
			TreeEntry: tree.Entries[i],
			TreeDepth: entry.TreeDepth + 1,
		}
		// If full when trying to append trees, then process non-tree entries
		if len(c.allEntries) == cap(c.allEntries) {
			// Make sure all go routines start processing non-tree entries
			if err = c.setNonTreeOnlyMode(); err != nil {
				return
			}
			// While channel is full handle the (non-tree) entries
			for len(c.allEntries) == cap(c.allEntries) {
				if err = c.handleEntry(<-c.allEntries); err != nil {
					return
				}
			}
		}
	}
	return
}

func (c *CloneObject) handleBlobEntry(entry Entry, totalSize *atomic.Int64, totalFileCount *atomic.Int64, trueSize *atomic.Int64) (err error) {
	blob, subErr := c.getBlobObject(entry.TreeEntry.Hash)
	if subErr != nil {
		err = subErr
		return
	}

	totalSize.Add(blob.Size)
	if totalSize.Load() > c.cfg.GetMaxTotalSize() {
		err = fmt.Errorf("%w: maximum repository size exceeded [%d > %d]", commonerrors.ErrTooLarge, totalSize.Load(), c.cfg.GetMaxTotalSize())
		return
	}

	if blob.Size > c.cfg.GetMaxFileSize() {
		err = fmt.Errorf("%w: maximum individual file size exceeded [%d > %d]", commonerrors.ErrTooLarge, blob.Size, c.cfg.GetMaxFileSize())
		return
	}

	if !c.seen.Contains(entry.TreeEntry.Hash) {
		c.seen.Add(entry.TreeEntry.Hash)
		trueSize.Add(blob.Size)
		if trueSize.Load() > c.cfg.GetMaxTrueSize() {
			err = fmt.Errorf("%w: maximum true size exceeded [%d > %d]", commonerrors.ErrTooLarge, trueSize.Load(), c.cfg.GetMaxTrueSize())
			return
		}
	}

	totalFileCount.Inc()
	if totalFileCount.Load() > c.cfg.GetMaxFileCount() {
		err = fmt.Errorf("%w: maximum file count exceeded [%d > %d]", commonerrors.ErrTooLarge, totalFileCount.Load(), c.cfg.GetMaxFileCount())
		return
	}
	c.resetNonTreeOnlyMode()
	return
}

func (c *CloneObject) handleCommitEntry(entry Entry) (err error) {
	// Unknown if necessary. Add code here if necessary in future
	c.resetNonTreeOnlyMode()
	return
}

func (c *CloneObject) handleSymlinkEntry(entry Entry) (err error) {
	// Unknown if necessary. Add code here if necessary in future
	c.resetNonTreeOnlyMode()
	return
}

func (c *CloneObject) checkDepthAndTotalEntries(entry Entry, totalEntries *atomic.Int64) (err error) {
	totalEntries.Inc()
	if totalEntries.Load() > c.cfg.GetMaxEntries() {
		err = fmt.Errorf("%w: maximum entries count exceeded [%d > %d]", commonerrors.ErrTooLarge, totalEntries.Load(), c.cfg.GetMaxEntries())
		return
	}

	if entry.TreeDepth > c.cfg.GetMaxTreeDepth() {
		err = fmt.Errorf("%w: maximum tree depth exceeded [%d > %d]", commonerrors.ErrTooLarge, entry.TreeDepth, c.cfg.GetMaxTreeDepth())
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
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		tree, subErr := trees.Next()
		if subErr != nil {
			if commonerrors.Any(subErr, io.EOF, commonerrors.ErrEOF) {
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
			entriesCount := len(c.allEntries)
			if entriesCount > int(c.cfg.GetMaxEntries()) {
				err = fmt.Errorf("%w: maximum entries count exceeded [%d >= %d]", commonerrors.ErrTooLarge, entriesCount, c.cfg.GetMaxEntries())
				return
			}
			if entriesCount == cap(c.allEntries) {
				err = fmt.Errorf("%w: entry channel saturated before initialisation complete [%d >= %d]", commonerrors.ErrTooLarge, entriesCount, MaxEntriesChannelSize)
				return
			}
		}
	}
	return
}

func (c *CloneObject) SetupLimits(cfg ILimits) (err error) {
	if cfg == nil {
		return fmt.Errorf("%w: limits config undefined", commonerrors.ErrUndefined)
	}
	c.cfg = cfg
	c.allEntries = make(chan Entry, MaxEntriesChannelSize)
	return
}

func NewCloneObject() *CloneObject {
	cloneObject := &CloneObject{
		cfg:                NoLimits(),
		totalSize:          atomic.NewInt64(0),
		totalFileCount:     atomic.NewInt64(0),
		totalEntries:       atomic.NewInt64(0),
		trueSize:           atomic.NewInt64(0),
		processNonTreeOnly: atomic.NewBool(false),
		treeSeenIdentifier: atomic.NewString(""),
		nonTreeOnlyMutex:   make(chan int, 1),
		seen:               mapset.NewSet[plumbing.Hash](),
	}
	cloneObject.nonTreeOnlyMutex <- 1
	return cloneObject
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

func (c *CloneObject) setNonTreeOnlyMode() (err error) {
	// If already set then do nothing and process existing round
	if c.processNonTreeOnly.Load() {
		return
	}
	// Generate a unique identifier so we can keep track of whether the trees have been seen before
	var seenIdentifier, subErr = idgen.GenerateUUID4()
	if subErr != nil {
		err = subErr
		return
	}
	// Launch as go func so this is none blocking
	// Also means that only the go routine that locked the resource can unlock it (so safer)
	go func() {
		select {
		case <-c.nonTreeOnlyMutex:
			c.treeSeenIdentifier.Store(seenIdentifier)
			c.processNonTreeOnly.Store(true)
			// Wait for processNonTreeOnly==false. This will happen when a non-tree entry is handled
			for c.processNonTreeOnly.Load() {
			}
			c.nonTreeOnlyMutex <- 1
		default:
			// If already locked then do nothing and process existing round
		}

		// TODO: when updated to go1.18 replace select with sync.Mutex and TryLock()
		/*
			// If already locked then do nothing and process existing round
			if c.nonTreeOnlyMutex.TryLock() {
				c.treeSeenIdentifier.Store(seenIdentifier)
				c.processNonTreeOnly.Store(true)
				// Wait for processNonTreeOnly==false. This will happen when a non-tree entry is handled
				for c.processNonTreeOnly.Load() {
				}
				c.nonTreeOnlyMutex.Unlock()
			}
		*/
	}()
	return
}

func (c *CloneObject) resetNonTreeOnlyMode() {
	c.processNonTreeOnly.Store(false)
}

// After cloning without checkout, valdiate the repository to check for git bombs
func (c *CloneObject) ValidateRepository(ctx context.Context) (err error) {
	if !c.cfg.Apply() {
		return
	}

	if err = c.populateInitialEntries(ctx); err != nil {
		return
	}

	errs, ctx := errgroup.WithContext(ctx)

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

				if err = c.checkDepthAndTotalEntries(entry, c.totalEntries); err != nil {
					return
				}

				if err = c.handleEntry(entry); err != nil {
					return
				}
			}
			return nil
		})
	}
	err = errs.Wait()
	return
}

func (c *CloneObject) parseReference(cfg *GitActionConfig) (branch plumbing.ReferenceName, hash plumbing.Hash) {
	ref := cfg.GetReference()
	if ref == "" {
		branch = plumbing.HEAD
		return
	}
	if strings.HasPrefix(ref, refPrefix) || strings.HasPrefix(ref, symrefPrefix) {
		branch = plumbing.ReferenceName(strings.TrimPrefix(ref, symrefPrefix))
	} else {
		if hashing.IsLikelyHexHashString(ref) {
			hash = plumbing.NewHash(ref)
		} else {
			tag, err := c.repo.Tag(ref)
			if err == nil {
				branch = tag.Name()
			} else {
				branch = plumbing.NewBranchReferenceName(ref)
			}
		}
	}
	return
}

func (c *CloneObject) Checkout(gitOptions *GitActionConfig) (err error) {
	worktree, err := c.repo.Worktree()
	if err != nil {
		return
	}
	branch, hash := c.parseReference(gitOptions)
	checkoutOptions := git.CheckoutOptions{
		Hash:   hash,
		Branch: branch,
		Create: gitOptions.GetCreate(),
	}
	return worktree.Checkout(&checkoutOptions)
}

// CloneWithLimits clones a repository with limits on the max tree depth, the max repository size, the max file count, the max individual file size, and the max entries
func CloneWithLimits(ctx context.Context, dir string, limits ILimits, gitOptions *GitActionConfig) (err error) {
	c := NewCloneObject()
	err = c.SetupLimits(limits)
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
