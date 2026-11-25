/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package git

import (
	"context"
	"iter"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	BranchRefPrefix = "refs/heads/"
	TagRefPrefix    = "refs/tags/"
	DefaultRemote   = "origin"
)

// ResolveReference returns a local commit hash without fetching or checkout.
// Short names prefer tags, then origin branches, then local branches.
// Call after Clone completes; cancellation is checked between Git reads.
// If the resolution was successful then revision will always have a value.
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
	case reflection.IsEmpty(reference) || reference == string(plumbing.HEAD):
		candidates = []plumbing.ReferenceName{plumbing.HEAD}
	case strings.HasPrefix(reference, BranchRefPrefix):
		branch := strings.TrimPrefix(reference, BranchRefPrefix)
		candidates = []plumbing.ReferenceName{plumbing.NewRemoteReferenceName(DefaultRemote, branch), plumbing.ReferenceName(reference)}
	case strings.HasPrefix(reference, TagRefPrefix):
		candidates = []plumbing.ReferenceName{plumbing.ReferenceName(reference)}
	case !strings.HasPrefix(reference, refPrefix):
		candidates = []plumbing.ReferenceName{plumbing.NewTagReferenceName(reference), plumbing.NewRemoteReferenceName(DefaultRemote, reference), plumbing.NewBranchReferenceName(reference)}
	}

	return
}

func (c *CloneObject) lookupReference(ctx context.Context, name plumbing.ReferenceName) (ref *plumbing.Reference, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	ref, err = c.getReference(name)
	if !commonerrors.Any(err, plumbing.ErrReferenceNotFound) {
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
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	for value, subErr := range c.referenceObjects(ctx, hash) {
		if subErr != nil {
			err = subErr
			return
		}

		if commit, ok := value.(*object.Commit); ok {
			commitHash = commit.Hash
			return
		}
	}

	err = commonerrors.New(commonerrors.ErrInvalid, "Git reference did not resolve to a commit")
	return
}

// referenceObjects yields a tag chain and its commit, or one terminal error.
func (c *CloneObject) referenceObjects(ctx context.Context, hash plumbing.Hash) iter.Seq2[object.Object, error] {
	return func(yield func(object.Object, error) bool) {
		seen := map[plumbing.Hash]bool{}
		objectType := plumbing.AnyObject

		for {
			if err := parallelisation.DetermineContextError(ctx); err != nil {
				yield(nil, err)
				return
			}

			if seen[hash] {
				yield(nil, commonerrors.Newf(commonerrors.ErrInvalid, "cyclic Git tag reference at %s", hash))
				return
			}
			seen[hash] = true

			value, err := c.getObject(objectType, hash)
			if err != nil {
				yield(nil, err)
				return
			}

			if err = parallelisation.DetermineContextError(ctx); err != nil {
				yield(nil, err)
				return
			}

			switch value := value.(type) {
			case *object.Commit:
				yield(value, nil)
				return
			case *object.Tag:
				if !yield(value, nil) {
					return
				}

				hash = value.Target
				objectType = value.TargetType
			default:
				yield(nil, commonerrors.Newf(commonerrors.ErrInvalid, "Git reference points to a %s, not a commit", value.Type()))
				return
			}
		}
	}
}
