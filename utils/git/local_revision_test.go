/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package git

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestLocalRevision(t *testing.T) {
	t.Parallel()

	t.Run("branch HEAD leaves staged, unstaged and untracked files unchanged", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		trackedFile := filepath.Join(fixture.directory, "tracked.txt")
		untrackedFile := filepath.Join(fixture.directory, "untracked.txt")
		require.NoError(t, filesystem.WriteFile(trackedFile, []byte("staged\n"), 0o600))
		worktree, err := fixture.repo.Worktree()
		require.NoError(t, err)
		_, err = worktree.Add("tracked.txt")
		require.NoError(t, err)
		require.NoError(t, filesystem.WriteFile(trackedFile, []byte("unstaged\n"), 0o600))
		require.NoError(t, filesystem.WriteFile(untrackedFile, []byte("untracked\n"), 0o600))

		nested := filepath.Join(fixture.directory, "nested", "directory")
		require.NoError(t, filesystem.MkDir(nested))
		before := revisionFiles(t, fixture.directory)

		for _, directory := range []string{fixture.directory, nested} {
			revision, subErr := LocalRevision(context.Background(), directory)
			require.NoError(t, subErr)
			require.NotNil(t, revision)
			assert.Equal(t, fixture.headCommit.String(), field.Optional(revision, ""))
		}

		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("detached HEAD leaves staged, unstaged and untracked files unchanged", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		head := plumbing.NewHashReference(plumbing.HEAD, fixture.headCommit)
		require.NoError(t, fixture.repo.Storer.SetReference(head))
		trackedFile := filepath.Join(fixture.directory, "tracked.txt")
		untrackedFile := filepath.Join(fixture.directory, "untracked.txt")
		require.NoError(t, filesystem.WriteFile(trackedFile, []byte("staged\n"), 0o600))
		worktree, err := fixture.repo.Worktree()
		require.NoError(t, err)
		_, err = worktree.Add("tracked.txt")
		require.NoError(t, err)
		require.NoError(t, filesystem.WriteFile(trackedFile, []byte("unstaged\n"), 0o600))
		require.NoError(t, filesystem.WriteFile(untrackedFile, []byte("untracked\n"), 0o600))

		nested := filepath.Join(fixture.directory, "nested", "directory")
		require.NoError(t, filesystem.MkDir(nested))
		before := revisionFiles(t, fixture.directory)

		for _, directory := range []string{fixture.directory, nested} {
			revision, subErr := LocalRevision(context.Background(), directory)
			require.NoError(t, subErr)
			require.NotNil(t, revision)
			assert.Equal(t, fixture.headCommit.String(), field.Optional(revision, ""))
		}

		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})
}

func TestLocalRevisionRelativePath(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	fs := filesystem.GetGlobalFileSystem()
	currentDirectory, err := fs.CurrentDirectory()
	require.NoError(t, err)
	relative, err := filesystem.FilePathRel(fs, currentDirectory, fixture.directory)
	require.NoError(t, err)

	revision, err := LocalRevision(context.Background(), relative)
	require.NoError(t, err)
	require.NotNil(t, revision)
	assert.Equal(t, fixture.headCommit.String(), field.Optional(revision, ""))
}

func TestLocalRevisionOptionalMetadata(t *testing.T) {
	t.Parallel()

	t.Run("ordinary directory", func(t *testing.T) {
		t.Parallel()

		directory := t.TempDir()
		before := revisionFiles(t, directory)

		revision, err := LocalRevision(context.Background(), directory)
		require.NoError(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, directory))
	})

	t.Run("unborn repository", func(t *testing.T) {
		t.Parallel()

		directory := t.TempDir()
		_, err := git.PlainInit(directory, false)
		require.NoError(t, err)
		before := revisionFiles(t, directory)

		revision, err := LocalRevision(context.Background(), directory)
		require.NoError(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, directory))
	})
}

func TestLocalRevisionLinkedWorktree(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	directory := t.TempDir()

	metadataDirectory := filepath.Join(".git", "worktrees", "linked")
	gitdir := filepath.Join(fixture.directory, metadataDirectory)
	require.NoError(t, filesystem.MkDir(gitdir))

	gitFile := filepath.Join(directory, ".git")
	gitFileContents := fmt.Sprintf("gitdir: %s\n", gitdir)
	require.NoError(t, filesystem.WriteFile(gitFile, []byte(gitFileContents), 0o600))
	commonDirFile := filepath.Join(gitdir, "commondir")
	backlinkFile := filepath.Join(gitdir, "gitdir")
	backlinkContents := fmt.Sprintf("%s\n", gitFile)
	headFile := filepath.Join(gitdir, "HEAD")
	require.NoError(t, filesystem.WriteFile(commonDirFile, []byte("../..\n"), 0o600))
	require.NoError(t, filesystem.WriteFile(backlinkFile, []byte(backlinkContents), 0o600))
	require.NoError(t, filesystem.WriteFile(headFile, []byte("ref: refs/heads/linked\n"), 0o600))
	linkedBranch := plumbing.NewBranchReferenceName("linked")
	linkedReference := plumbing.NewHashReference(linkedBranch, fixture.initialCommit)
	require.NoError(t, fixture.repo.Storer.SetReference(linkedReference))

	nested := filepath.Join(directory, "nested")
	require.NoError(t, filesystem.MkDir(nested))
	beforeMain := revisionFiles(t, fixture.directory)
	beforeLinked := revisionFiles(t, directory)

	for _, path := range []string{directory, nested} {
		revision, err := LocalRevision(context.Background(), path)
		require.NoError(t, err)
		require.NotNil(t, revision)
		assert.Equal(t, fixture.initialCommit.String(), field.Optional(revision, ""))
	}

	assert.Equal(t, beforeMain, revisionFiles(t, fixture.directory))
	assert.Equal(t, beforeLinked, revisionFiles(t, directory))
}

func TestLocalRevisionInvalidDirectories(t *testing.T) {
	t.Parallel()

	fixture := newRevisionFixture(t)
	before := revisionFiles(t, fixture.directory)

	t.Run("missing directory", func(t *testing.T) {
		directory := filepath.Join(fixture.directory, "missing")

		revision, err := LocalRevision(context.Background(), directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("file instead of directory", func(t *testing.T) {
		filename := filepath.Join(fixture.directory, "tracked.txt")

		revision, err := LocalRevision(context.Background(), filename)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})
}

func TestLocalRevisionErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid git file", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		gitFile := filepath.Join(fixture.directory, ".git")
		require.NoError(t, filesystem.Rm(gitFile))
		require.NoError(t, filesystem.WriteFile(gitFile, []byte("invalid\n"), 0o600))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("missing HEAD", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		headFile := filepath.Join(fixture.directory, ".git", "HEAD")
		require.NoError(t, filesystem.Rm(headFile))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("corrupt HEAD", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		headFile := filepath.Join(fixture.directory, ".git", "HEAD")
		require.NoError(t, filesystem.WriteFile(headFile, []byte("invalid\n"), 0o600))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("HEAD is a blob", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		head := plumbing.NewHashReference(plumbing.HEAD, fixture.blobHash)
		require.NoError(t, fixture.repo.Storer.SetReference(head))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("HEAD commit missing", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		head := plumbing.NewHashReference(plumbing.HEAD, plumbing.ZeroHash)
		require.NoError(t, fixture.repo.Storer.SetReference(head))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("unborn with corrupt packed refs", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		unbornBranch := plumbing.NewBranchReferenceName("unborn")
		head := plumbing.NewSymbolicReference(plumbing.HEAD, unbornBranch)
		require.NoError(t, fixture.repo.Storer.SetReference(head))
		packedRefsFile := filepath.Join(fixture.directory, ".git", "packed-refs")
		require.NoError(t, filesystem.WriteFile(packedRefsFile, []byte("invalid\n"), 0o600))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("unborn with empty loose ref", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		unbornBranch := plumbing.NewBranchReferenceName("unborn")
		head := plumbing.NewSymbolicReference(plumbing.HEAD, unbornBranch)
		require.NoError(t, fixture.repo.Storer.SetReference(head))
		emptyRefFile := filepath.Join(fixture.directory, ".git", "refs", "heads", "empty")
		emptyRef, err := filesystem.CreateFile(emptyRefFile)
		require.NoError(t, err)
		require.NoError(t, emptyRef.Close())
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})

	t.Run("invalid symbolic HEAD", func(t *testing.T) {
		t.Parallel()

		fixture := newRevisionFixture(t)
		headFile := filepath.Join(fixture.directory, ".git", "HEAD")
		require.NoError(t, filesystem.WriteFile(headFile, []byte("ref: refs/heads/invalid~name\n"), 0o600))
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(context.Background(), fixture.directory)
		require.Error(t, err)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})
}

func TestLocalRevisionInaccessibleMetadata(t *testing.T) {
	t.Parallel()

	for _, path := range []string{".", ".git", ".git/refs/heads"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			fixture := newRevisionFixture(t)

			// A missing loose branch must not conceal unreadable refs as unborn.
			unbornBranch := plumbing.NewBranchReferenceName("unborn")
			head := plumbing.NewSymbolicReference(plumbing.HEAD, unbornBranch)
			require.NoError(t, fixture.repo.Storer.SetReference(head))
			before := revisionFiles(t, fixture.directory)
			inaccessible := filepath.Join(fixture.directory, filepath.FromSlash(path))
			info, err := filesystem.Stat(inaccessible)
			require.NoError(t, err)
			require.NoError(t, filesystem.Chmod(inaccessible, 0))
			t.Cleanup(func() {
				require.NoError(t, filesystem.Chmod(inaccessible, info.Mode().Perm()))
			})

			mainRefFile := filepath.Join(fixture.directory, ".git", "refs", "heads", "main")
			_, err = filesystem.ReadFile(mainRefFile)
			if err == nil {
				t.Skip("directory permissions are not enforced for this user or filesystem")
			}

			revision, err := LocalRevision(context.Background(), fixture.directory)
			require.Error(t, err)
			assert.Nil(t, revision)
			require.NoError(t, filesystem.Chmod(inaccessible, info.Mode().Perm()))
			assert.Equal(t, before, revisionFiles(t, fixture.directory))
		})
	}
}

func TestLocalRevisionCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("cancellation takes precedence over a missing directory", func(t *testing.T) {
		directory := filepath.Join(t.TempDir(), "missing")

		revision, err := LocalRevision(ctx, directory)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.Nil(t, revision)
	})

	t.Run("cancelled discovery leaves the repository unchanged", func(t *testing.T) {
		fixture := newRevisionFixture(t)
		before := revisionFiles(t, fixture.directory)

		revision, err := LocalRevision(ctx, fixture.directory)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.Nil(t, revision)
		assert.Equal(t, before, revisionFiles(t, fixture.directory))
	})
}
