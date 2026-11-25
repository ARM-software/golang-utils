/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package git

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// LocalRevision reads HEAD without changing the worktree or index.
// Non-Git directories and valid unborn HEADs return an empty revision.
func LocalRevision(ctx context.Context, directory string) (revision string, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	directory, err = filepath.Abs(directory)
	if err != nil {
		return
	}

	info, err := filesystem.Stat(directory)
	if err != nil {
		return
	}

	if !info.IsDir() {
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
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		err = validateUnbornLocalHEAD(ctx, repo)
		return
	}

	if err != nil {
		return
	}

	if _, err = repo.CommitObject(head.Hash()); err != nil {
		return
	}

	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	revision = head.Hash().String()
	return
}

func hasLocalGitMetadata(ctx context.Context, directory string) (found bool, err error) {
	// Absence is optional, unreadable metadata must not be treated as absence.
	for ancestor := directory; ; ancestor = filepath.Dir(ancestor) {
		if err = parallelisation.DetermineContextError(ctx); err != nil {
			return
		}

		_, subErr := filesystem.Lstat(filepath.Join(ancestor, ".git"))
		if subErr == nil {
			found = true
			return
		}

		if !commonerrors.Any(subErr, commonerrors.ErrNotFound) {
			err = subErr
			return
		}

		if filepath.Dir(ancestor) == ancestor {
			return
		}
	}
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
		return
	}

	err = parallelisation.DetermineContextError(ctx)
	return
}
