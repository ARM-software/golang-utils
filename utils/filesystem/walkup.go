/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// WalkUp traverses directories upwards from start to the filesystem root.
//
// The traversal is lexical: it resolves start to an absolute path without
// resolving symbolic links, visits the starting directory first, then each of
// its ancestors, and finally the root exactly once. If fn returns stop=true,
// traversal stops without error. If fn returns an error, traversal stops and
// that error is returned. Context cancellation is checked before traversal
// starts and before each callback invocation.
func WalkUp(ctx context.Context, start string, fn func(dir string) (stop bool, err error)) error {
	return GetGlobalFileSystem().WalkUp(ctx, start, fn)
}

// WalkUpWithExclusionPatterns traverses directories upwards from start to the
// filesystem root while skipping callback invocations for directories matching
// any exclusion pattern.
//
// Exclusion patterns use the same regex semantics as the other filesystem
// exclusion helpers. Matching directories are skipped lexically and traversal
// continues with their ancestors. Context cancellation is checked before
// traversal starts and before each non-excluded callback invocation.
func WalkUpWithExclusionPatterns(ctx context.Context, start string, fn func(dir string) (stop bool, err error), exclusionPatterns ...string) error {
	return GetGlobalFileSystem().WalkUpWithExclusionPatterns(ctx, start, fn, exclusionPatterns...)
}

// FindUp searches for pattern from start upwards to the filesystem root.
//
// The search is lexical and checks the starting directory before its
// ancestors. The returned path is absolute. pattern must be relative to the
// visited directory and must not escape it; absolute patterns,
// volume-qualified patterns, and patterns beginning with `..` are rejected
// with commonerrors.ErrInvalid. Pattern matching uses the filesystem Glob
// implementation, including doublestar support. If no match is found,
// commonerrors.ErrNotFound is returned. Context cancellation and deadline
// errors are propagated as-is.
func FindUp(ctx context.Context, start, pattern string) (string, error) {
	return GetGlobalFileSystem().FindUp(ctx, start, pattern)
}

// WalkUp traverses directories upwards from start to the filesystem root.
//
// The traversal is lexical: it resolves start to an absolute path without
// resolving symbolic links, visits the starting directory first, then each of
// its ancestors, and finally the root exactly once. If fn returns stop=true,
// traversal stops without error. If fn returns an error, traversal stops and
// that error is returned. Context cancellation is checked before traversal
// starts and before each callback invocation.
func (fs *VFS) WalkUp(ctx context.Context, start string, fn func(dir string) (stop bool, err error)) error {
	return fs.WalkUpWithExclusionPatterns(ctx, start, fn)
}

// WalkUpWithExclusionPatterns traverses directories upwards from start to the
// filesystem root while skipping callback invocations for directories matching
// any exclusion pattern.
//
// Exclusion patterns use the same regex semantics as the other filesystem
// exclusion helpers. Matching directories are skipped lexically and traversal
// continues with their ancestors. Context cancellation is checked before
// traversal starts and before each non-excluded callback invocation.
func (fs *VFS) WalkUpWithExclusionPatterns(ctx context.Context, start string, fn func(dir string) (stop bool, err error), exclusionPatterns ...string) error {
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return err
	}
	if fn == nil {
		return commonerrors.UndefinedVariable("walk function")
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	currentDirectory, err := fs.CurrentDirectory()
	if err != nil {
		return err
	}
	exclusionRegex, err := NewExclusionRegexList(fs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return err
	}
	startAbs := FilePathClean(fs, FilePathAbs(fs, start, currentDirectory))
	startInfo, err := fs.Stat(startAbs)
	if err != nil {
		return ConvertFileSystemError(err)
	}
	if !startInfo.IsDir() {
		return commonerrors.Newf(commonerrors.ErrInvalid, "path [%v] is not a directory", startAbs)
	}
	return walkUpLexically(ctx, fs, startAbs, exclusionRegex, fn)
}

// FindUp searches for pattern from start upwards to the filesystem root.
//
// The search is lexical and checks the starting directory before its
// ancestors. The returned path is absolute. pattern must be relative to the
// visited directory and must not escape it; absolute patterns,
// volume-qualified patterns, and patterns beginning with `..` are rejected
// with commonerrors.ErrInvalid. Pattern matching uses the filesystem Glob
// implementation, including doublestar support. If no match is found,
// commonerrors.ErrNotFound is returned. Context cancellation and deadline
// errors are propagated as-is.
func (fs *VFS) FindUp(ctx context.Context, start, pattern string) (string, error) {
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return "", err
	}
	err = validateFindUpPattern(fs, pattern)
	if err != nil {
		return "", err
	}
	var found string
	err = fs.WalkUp(ctx, start, func(dir string) (bool, error) {
		matches, matchErr := fs.Glob(FilePathJoin(fs, dir, pattern))
		switch convertedErr := ConvertFileSystemError(matchErr); {
		case convertedErr == nil:
			if len(matches) == 0 {
				return false, nil
			}
			sort.Strings(matches)
			found = matches[0]
			return true, nil
		case commonerrors.Any(convertedErr, commonerrors.ErrNotFound):
			return false, nil
		default:
			return false, convertedErr
		}
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", commonerrors.Newf(commonerrors.ErrNotFound, "could not find pattern [%v] from [%v]", pattern, start)
	}
	return found, nil
}

func walkUpLexically(ctx context.Context, fs FS, start string, exclusions []*regexp.Regexp, fn func(dir string) (stop bool, err error)) error {
	current := FilePathClean(fs, start)
	for {
		if !IsPathExcluded(current, exclusions...) {
			err := parallelisation.DetermineContextError(ctx)
			if err != nil {
				return err
			}
			stop, err := fn(current)
			if err != nil || stop {
				return err
			}
		}
		parent := FilePathDir(fs, current)
		if parent == current {
			return nil
		}
		current = parent
	}
}

func validateFindUpPattern(fs FS, pattern string) error {
	if pattern == "" {
		return commonerrors.UndefinedVariable("search path")
	}
	if FilePathIsAbs(fs, pattern) || FilePathVolumeName(fs, pattern) != "" {
		return commonerrors.Newf(commonerrors.ErrInvalid, "search path [%v] must be relative", pattern)
	}
	cleanPattern := FilePathClean(fs, pattern)
	prefix := ".." + string(fs.PathSeparator())
	if cleanPattern == ".." || strings.HasPrefix(cleanPattern, prefix) {
		return commonerrors.Newf(commonerrors.ErrInvalid, "search path [%v] must not escape the starting directory", pattern)
	}
	return nil
}
