/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package git

import (
	"context"
	"errors"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	branchRefPrefix = "refs/heads/"
	tagRefPrefix    = "refs/tags/"
	defaultRemote   = "origin"
)

// ResolveReference returns a local commit hash without fetching or checkout.
// Short names prefer tags, then origin branches, then local branches.
// Call after Clone completes; cancellation is checked between Git reads.
func (c *CloneObject) ResolveReference(ctx context.Context, reference string) (revision string, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	if c == nil || c.repo == nil {
		err = commonerrors.UndefinedVariable("cloned repository")
		return
	}

	for _, candidate := range referenceCandidates(reference) {
		ref, subErr := c.lookupReference(ctx, candidate)
		if subErr != nil {
			err = subErr
			return
		}

		if ref == nil {
			continue
		}

		hash, subErr := c.peelReferenceCommit(ctx, ref.Hash())
		if subErr != nil {
			err = subErr
			return
		}

		revision = hash.String()
		return
	}

	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	if !plumbing.IsHash(reference) {
		err = commonerrors.Newf(commonerrors.ErrNotFound, "Git reference %q is not available locally", reference)
		return
	}

	hash := plumbing.NewHash(reference)
	if _, err = c.getCommitObject(hash); err != nil {
		return
	}

	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	revision = hash.String()
	return
}

func referenceCandidates(reference string) (candidates []plumbing.ReferenceName) {
	switch {
	case reference == "" || reference == string(plumbing.HEAD):
		candidates = []plumbing.ReferenceName{plumbing.HEAD}
	case strings.HasPrefix(reference, branchRefPrefix):
		branch := strings.TrimPrefix(reference, branchRefPrefix)
		candidates = []plumbing.ReferenceName{plumbing.NewRemoteReferenceName(defaultRemote, branch), plumbing.ReferenceName(reference)}
	case strings.HasPrefix(reference, tagRefPrefix):
		candidates = []plumbing.ReferenceName{plumbing.ReferenceName(reference)}
	case !strings.HasPrefix(reference, refPrefix):
		candidates = []plumbing.ReferenceName{plumbing.NewTagReferenceName(reference), plumbing.NewRemoteReferenceName(defaultRemote, reference), plumbing.NewBranchReferenceName(reference)}
	}

	return
}

func (c *CloneObject) lookupReference(ctx context.Context, name plumbing.ReferenceName) (ref *plumbing.Reference, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	ref, err = c.getReference(name)
	if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return
	}

	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	// Loose-ref lookup can hide read errors; verify absence before falling back.
	err = c.checkReferences()

	return
}

func (c *CloneObject) peelReferenceCommit(ctx context.Context, hash plumbing.Hash) (commitHash plumbing.Hash, err error) {
	objectType := plumbing.AnyObject

	for {
		if err = parallelisation.DetermineContextError(ctx); err != nil {
			return
		}

		value, subErr := c.getObject(objectType, hash)
		if subErr != nil {
			err = subErr
			return
		}

		switch value := value.(type) {
		case *object.Commit:
			if err = parallelisation.DetermineContextError(ctx); err != nil {
				return
			}

			commitHash = value.Hash
			return
		case *object.Tag:
			hash = value.Target
			objectType = value.TargetType
		default:
			err = commonerrors.Newf(commonerrors.ErrInvalid, "Git reference points to a %s, not a commit", value.Type())
			return
		}
	}
}
