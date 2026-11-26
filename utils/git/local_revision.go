/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package git

import (
	"context"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// LocalRevision reads HEAD without changing the worktree or index.
// Non-Git directories and valid unborn HEADs return nil.
func LocalRevision(ctx context.Context, directory string) (revision *string, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	fs := filesystem.GetGlobalFileSystem()
	var currentDirectory string
	if !filesystem.FilePathIsAbs(fs, directory) {
		currentDirectory, err = fs.CurrentDirectory()
		if err != nil {
			return
		}
	}
	directory = filesystem.FilePathAbs(fs, directory, currentDirectory)

	info, err := fs.Stat(directory)
	if err != nil {
		return
	}

	if !filesystem.IsDirectory(info) {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "not a directory: %q", directory)
		return
	}

	found, err := hasLocalGitMetadata(ctx, directory)
	if err != nil || !found {
		return
	}

	repo, err := git.PlainOpenWithOptions(directory, &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return
	}

	head, err := repo.Head()
	if commonerrors.Any(err, plumbing.ErrReferenceNotFound) {
		err = validateUnbornLocalHEAD(ctx, repo)
		return
	}

	if err != nil {
		return
	}

	if _, err = repo.CommitObject(head.Hash()); err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "could not read local HEAD commit")
		return
	}

	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	revision = field.ToOptional(head.Hash().String())
	return
}

func hasLocalGitMetadata(ctx context.Context, directory string) (found bool, err error) {
	fs := filesystem.GetGlobalFileSystem()

	err = fs.WalkUp(ctx, directory, func(ancestor string) (stop bool, subErr error) {
		// Use Lstat to distinguish absent metadata from inspection errors.
		_, subErr = fs.Lstat(filesystem.FilePathJoin(fs, ancestor, ".git"))
		if subErr == nil {
			found = true
			stop = true
			return
		}

		subErr = commonerrors.Ignore(subErr, commonerrors.ErrNotFound)
		return
	})

	return
}

func validateUnbornLocalHEAD(ctx context.Context, repo *git.Repository) (err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	// Enumerate refs to expose read failures hidden by loose-ref lookup.
	refs, err := repo.References()
	if err != nil {
		return
	}
	refs.Close()

	head, err := repo.Reference(plumbing.HEAD, false)
	if err != nil {
		return
	}

	if head.Type() != plumbing.SymbolicReference {
		err = commonerrors.New(commonerrors.ErrInvalid, "invalid local Git HEAD")
		return
	}

	if err = head.Target().Validate(); err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid local Git HEAD target")
		return
	}

	err = parallelisation.DetermineContextError(ctx)
	return
}
