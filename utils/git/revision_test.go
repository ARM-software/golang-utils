/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveReferenceHEAD(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	t.Run("empty reference selects HEAD", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "")
		require.NoError(t, err)
		assert.Equal(t, fixture.headCommit.String(), revision)
	})

	t.Run("explicit HEAD", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "HEAD")
		require.NoError(t, err)
		assert.Equal(t, fixture.headCommit.String(), revision)
	})

	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceBranches(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	directory := t.TempDir()
	clone := NewCloneObject()
	require.NoError(t, clone.Clone(context.Background(), directory, &GitActionConfig{
		URL: fixture.directory, NoCheckout: true, Tags: git.AllTags,
	}))

	// Keep origin/main at the latest commit, but move local main to the first.
	localMain := plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), fixture.initialCommit)
	localOnly := plumbing.NewHashReference(plumbing.NewBranchReferenceName("local-only"), fixture.initialCommit)
	require.NoError(t, clone.repo.Storer.SetReference(localMain))
	require.NoError(t, clone.repo.Storer.SetReference(localOnly))
	before := revisionFiles(t, directory)

	t.Run("HEAD follows the local branch", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "HEAD")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("short branch prefers origin", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "main")
		require.NoError(t, err)
		assert.Equal(t, fixture.headCommit.String(), revision)
	})

	t.Run("qualified branch prefers origin", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/heads/main")
		require.NoError(t, err)
		assert.Equal(t, fixture.headCommit.String(), revision)
	})

	t.Run("short branch falls back to local", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "local-only")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("qualified branch falls back to local", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/heads/local-only")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	assert.Equal(t, before, revisionFiles(t, directory))
}

func TestResolveReferenceTagBranchCollision(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	branch := plumbing.NewHashReference(plumbing.NewRemoteReferenceName("origin", "shared"), fixture.headCommit)
	require.NoError(t, fixture.repo.Storer.SetReference(branch))
	_, err := fixture.repo.CreateTag("shared", fixture.initialCommit, nil)
	require.NoError(t, err)
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	t.Run("short name selects tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "shared")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("qualified tag selects tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/tags/shared")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("qualified branch bypasses tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/heads/shared")
		require.NoError(t, err)
		assert.Equal(t, fixture.headCommit.String(), revision)
	})

	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceAnnotatedTags(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	tagger := &object.Signature{Name: "Revision Test", Email: "revision@example.invalid", When: time.Unix(0, 0)}
	annotated, err := fixture.repo.CreateTag("annotated", fixture.initialCommit, &git.CreateTagOptions{
		Tagger: tagger, Message: "Points to a commit",
	})
	require.NoError(t, err)
	nested, err := fixture.repo.CreateTag("nested", annotated.Hash(), &git.CreateTagOptions{
		Tagger: tagger, Message: "Points to another tag",
	})
	require.NoError(t, err)
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	t.Run("short annotated tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "annotated")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("qualified annotated tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/tags/annotated")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("short nested tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "nested")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("qualified nested tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/tags/nested")
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("raw tag hash is not a commit hash", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), annotated.Hash().String())
		require.Error(t, err)
		assert.Empty(t, revision)
	})

	t.Run("raw nested-tag hash is not a commit hash", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), nested.Hash().String())
		require.Error(t, err)
		assert.Empty(t, revision)
	})

	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceCommitHashes(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	t.Run("historical commit", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), fixture.initialCommit.String())
		require.NoError(t, err)
		assert.Equal(t, fixture.initialCommit.String(), revision)
	})

	t.Run("upper-case commit hash", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), strings.ToUpper(fixture.headCommit.String()))
		require.NoError(t, err)
		assert.Equal(t, fixture.headCommit.String(), revision)
	})

	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceEmptyCollidingTag(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	branch := plumbing.NewHashReference(plumbing.NewBranchReferenceName("shared"), fixture.headCommit)
	remote := plumbing.NewHashReference(plumbing.NewRemoteReferenceName("origin", "shared"), fixture.headCommit)
	require.NoError(t, fixture.repo.Storer.SetReference(branch))
	require.NoError(t, fixture.repo.Storer.SetReference(remote))
	_, err := fixture.repo.CreateTag("shared", fixture.initialCommit, nil)
	require.NoError(t, err)

	tagFile := filepath.Join(fixture.directory, ".git", "refs", "tags", "shared")
	require.NoError(t, filesystem.WriteFile(tagFile, nil, 0o600))
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	revision, err := clone.ResolveReference(context.Background(), "shared")
	require.Error(t, err, "a corrupt tag must not fall back to the colliding branch")
	assert.Empty(t, revision)
	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceMissingNames(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	t.Run("short name", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "missing")
		errortest.RequireError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, revision)
	})

	t.Run("qualified branch", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/heads/missing")
		errortest.RequireError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, revision)
	})

	t.Run("qualified tag", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/tags/missing")
		errortest.RequireError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, revision)
	})

	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceUnsupportedForms(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	clone := &CloneObject{repo: fixture.repo}
	before := revisionFiles(t, fixture.directory)

	t.Run("remote namespace", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "refs/remotes/origin/main")
		require.Error(t, err)
		assert.Empty(t, revision)
	})

	t.Run("revision expression", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), "HEAD~1")
		require.Error(t, err)
		assert.Empty(t, revision)
	})

	t.Run("abbreviated hash", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), fixture.initialCommit.String()[:12])
		require.Error(t, err)
		assert.Empty(t, revision)
	})

	t.Run("missing commit object", func(t *testing.T) {
		revision, err := clone.ResolveReference(context.Background(), strings.Repeat("0", 40))
		require.Error(t, err)
		assert.Empty(t, revision)
	})

	assert.Equal(t, before, revisionFiles(t, fixture.directory))
}

func TestResolveReferenceNonCommitObjects(t *testing.T) {
	t.Parallel()

	t.Run("raw blob hash", func(t *testing.T) {
		fixture := newRevisionFixture(t)
		clone := &CloneObject{repo: fixture.repo}
		before := revisionFiles(t, fixture.directory)

		revision, err := clone.ResolveReference(context.Background(), fixture.blobHash.String())
		require.Error(t, err)
		assert.Empty(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("lightweight tag pointing to a blob", func(t *testing.T) {
		fixture := newRevisionFixture(t)
		tag := plumbing.NewHashReference(plumbing.NewTagReferenceName("blob"), fixture.blobHash)
		require.NoError(t, fixture.repo.Storer.SetReference(tag))
		clone := &CloneObject{repo: fixture.repo}
		before := revisionFiles(t, fixture.directory)

		revision, err := clone.ResolveReference(context.Background(), "blob")
		require.Error(t, err)
		assert.Empty(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("tag pointing to a missing object", func(t *testing.T) {
		fixture := newRevisionFixture(t)
		tag := plumbing.NewHashReference(plumbing.NewTagReferenceName("broken"), plumbing.ZeroHash)
		require.NoError(t, fixture.repo.Storer.SetReference(tag))
		clone := &CloneObject{repo: fixture.repo}
		before := revisionFiles(t, fixture.directory)

		revision, err := clone.ResolveReference(context.Background(), "broken")
		require.Error(t, err)
		assert.Empty(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("annotated tag pointing to a blob", func(t *testing.T) {
		fixture := newRevisionFixture(t)
		tag, err := fixture.repo.CreateTag("annotated-blob", fixture.blobHash, &git.CreateTagOptions{
			Tagger:  &object.Signature{Name: "Revision Test", Email: "revision@example.invalid", When: time.Unix(0, 0)},
			Message: "Not a commit",
		})
		require.NoError(t, err)
		clone := &CloneObject{repo: fixture.repo}
		before := revisionFiles(t, fixture.directory)

		revision, err := clone.ResolveReference(context.Background(), "annotated-blob")
		require.Error(t, err)
		assert.Empty(t, revision)

		revision, err = clone.ResolveReference(context.Background(), tag.Hash().String())
		require.Error(t, err)
		assert.Empty(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})
}

func TestResolveReferenceBeforeClone(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		revision, err := (*CloneObject)(nil).ResolveReference(context.Background(), "HEAD")
		errortest.RequireError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, revision)
	})

	t.Run("zero-value receiver", func(t *testing.T) {
		clone := &CloneObject{}
		revision, err := clone.ResolveReference(context.Background(), "HEAD")
		errortest.RequireError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, revision)
	})

	t.Run("initialised but not cloned", func(t *testing.T) {
		clone := NewCloneObject()
		revision, err := clone.ResolveReference(context.Background(), "HEAD")
		errortest.RequireError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, revision)
	})
}

func TestResolveReferenceConcurrentReads(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	clone := &CloneObject{repo: fixture.repo}
	references := []string{"HEAD", "main", fixture.headCommit.String()}
	const readersPerReference = 4
	readers := len(references) * readersPerReference

	type resolutionResult struct {
		revision string
		err      error
	}
	results := make(chan resolutionResult, readers)

	for _, reference := range references {
		for range readersPerReference {
			go func() {
				revision, err := clone.ResolveReference(context.Background(), reference)
				results <- resolutionResult{revision: revision, err: err}
			}()
		}
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	for range readers {
		select {
		case resolved := <-results:
			require.NoError(t, resolved.err)
			assert.Equal(t, fixture.headCommit.String(), resolved.revision)
		case <-timer.C:
			require.FailNow(t, "concurrent revision reads did not complete")
		}
	}
}

func TestResolveReferenceCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("before a repository exists", func(t *testing.T) {
		revision, err := (*CloneObject)(nil).ResolveReference(ctx, "HEAD")
		errortest.AssertError(t, err, context.Canceled)
		assert.Empty(t, revision)
	})

	t.Run("existing repository", func(t *testing.T) {
		fixture := newRevisionFixture(t)
		clone := &CloneObject{repo: fixture.repo}
		before := revisionFiles(t, fixture.directory)

		revision, err := clone.ResolveReference(ctx, "main")
		errortest.AssertError(t, err, context.Canceled)
		assert.Empty(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})
}

type revisionFixture struct {
	repo          *git.Repository
	directory     string
	initialCommit plumbing.Hash
	headCommit    plumbing.Hash
	blobHash      plumbing.Hash
}

func newRevisionFixture(t *testing.T) (fixture revisionFixture) {
	t.Helper()

	fixture.directory = t.TempDir()
	repo, err := git.PlainInit(fixture.directory, false)
	require.NoError(t, err)
	fixture.repo = repo
	main := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main"))
	require.NoError(t, repo.Storer.SetReference(main))

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	filename := filepath.Join(fixture.directory, "tracked.txt")

	for i, content := range []string{"first\n", "second\n"} {
		require.NoError(t, filesystem.WriteFile(filename, []byte(content), 0o600))
		_, err = worktree.Add("tracked.txt")
		require.NoError(t, err)

		hash, subErr := worktree.Commit(content, &git.CommitOptions{Author: &object.Signature{
			Name: "Revision Test", Email: "revision@example.invalid", When: time.Unix(int64(i), 0),
		}})
		require.NoError(t, subErr)

		if i == 0 {
			fixture.initialCommit = hash
		} else {
			fixture.headCommit = hash
		}
	}

	commit, err := repo.CommitObject(fixture.headCommit)
	require.NoError(t, err)
	file, err := commit.File("tracked.txt")
	require.NoError(t, err)
	fixture.blobHash = file.Hash
	return
}

// Include metadata and file contents to detect writes; exclude access times.
func revisionFiles(t *testing.T, directory string) (files map[string]string) {
	t.Helper()

	var paths []string
	require.NoError(t, filesystem.ListDirTree(directory, &paths))
	files = make(map[string]string)

	for _, path := range paths {
		info, err := filesystem.Lstat(path)
		require.NoError(t, err)

		var content []byte
		if !info.IsDir() {
			content, err = filesystem.ReadFile(path)
			require.NoError(t, err)
		}

		files[path] = fmt.Sprintf("%s %s\n%s", info.Mode(), info.ModTime(), content)
	}

	return
}
