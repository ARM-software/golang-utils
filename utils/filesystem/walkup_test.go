/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestWalkUp(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup")
			start := FilePathJoin(fs, root, "a", "b", "c")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			var visited []string
			err := fs.WalkUp(context.Background(), start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				return false, nil
			})
			require.NoError(t, err)
			assert.Equal(t, expectedWalkUpOrder(t, fs, start), visited)
		})
	}
}

func TestWalkUp_RelativeStart(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			currentDir, err := fs.CurrentDirectory()
			require.NoError(t, err)
			rootName := fmt.Sprintf(".walkup-relative-%s-%v", sanitiseTestName(t.Name()), fsType)
			start := FilePathJoin(fs, rootName, "child")
			absStart := FilePathJoin(fs, currentDir, start)
			defer func() { _ = fs.Rm(FilePathJoin(fs, currentDir, rootName)) }()
			require.NoError(t, fs.MkDirAll(absStart, 0o755))

			var visited []string
			err = fs.WalkUp(context.Background(), start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				return false, nil
			})
			require.NoError(t, err)
			assert.Equal(t, expectedWalkUpOrder(t, fs, absStart), visited)
		})
	}
}

func TestWalkUp_StopEarly(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-stop")
			start := FilePathJoin(fs, root, "a", "b", "c")
			require.NoError(t, fs.MkDirAll(start, 0o755))
			stopAt := FilePathDir(fs, start)

			var visited []string
			err := fs.WalkUp(context.Background(), start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				return dir == stopAt, nil
			})
			require.NoError(t, err)
			assert.Equal(t, []string{start, stopAt}, visited)
		})
	}
}

func TestWalkUpWithExclusionPatterns(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-exclusion")
			start := FilePathJoin(fs, root, "a", "skip-me", "c")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			var visited []string
			err := fs.WalkUpWithExclusionPatterns(context.Background(), start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				return false, nil
			}, `.*skip-me.*`)
			require.NoError(t, err)
			assert.Equal(t, []string{FilePathJoin(fs, root, "a"), root}, visited[:2])
			assert.NotContains(t, visited, start)
			assert.NotContains(t, visited, FilePathJoin(fs, root, "a", "skip-me"))
		})
	}
}

func TestWalkUpWithExclusionPatterns_ExcludedStartStillWalksAncestors(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-exclusion-start")
			start := FilePathJoin(fs, root, "excluded")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			var visited []string
			err := fs.WalkUpWithExclusionPatterns(context.Background(), start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				return false, nil
			}, `.*excluded$`)
			require.NoError(t, err)
			assert.NotContains(t, visited, start)
			assert.Contains(t, visited, root)
		})
	}
}

func TestWalkUp_CallbackError(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-error")
			start := FilePathJoin(fs, root, "a", "b")
			require.NoError(t, fs.MkDirAll(start, 0o755))
			expectedErr := errors.New("stop")

			err := fs.WalkUp(context.Background(), start, func(dir string) (bool, error) {
				if dir == start {
					return false, expectedErr
				}
				return false, nil
			})
			require.ErrorIs(t, err, expectedErr)
		})
	}
}

func TestWalkUp_InvalidStart(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-invalid")

			missing := FilePathJoin(fs, root, "missing")
			err := fs.WalkUp(context.Background(), missing, func(dir string) (bool, error) {
				return false, nil
			})
			errortest.AssertError(t, err, commonerrors.ErrNotFound)

			filePath := FilePathJoin(fs, root, "file.txt")
			writeWalkUpTestFile(t, fs, filePath)
			err = fs.WalkUp(context.Background(), filePath, func(dir string) (bool, error) {
				return false, nil
			})
			errortest.AssertError(t, err, commonerrors.ErrInvalid)
		})
	}
}

func TestWalkUp_ContextErrors(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("cancelled_before_start_for_fs_%v", fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-cancel")
			start := FilePathJoin(fs, root, "a", "b")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			called := false
			err := fs.WalkUp(ctx, start, func(dir string) (bool, error) {
				called = true
				return false, nil
			})
			assert.False(t, called)
			errortest.AssertError(t, err, commonerrors.ErrCancelled)
		})

		t.Run(fmt.Sprintf("cancelled_during_traversal_for_fs_%v", fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-cancel-during")
			start := FilePathJoin(fs, root, "a", "b", "c")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var visited []string
			err := fs.WalkUp(ctx, start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				cancel()
				return false, nil
			})
			assert.Equal(t, []string{start}, visited)
			errortest.AssertError(t, err, commonerrors.ErrCancelled)
		})

		t.Run(fmt.Sprintf("deadline_exceeded_for_fs_%v", fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-deadline")
			start := FilePathJoin(fs, root, "a")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			defer cancel()
			err := fs.WalkUp(ctx, start, func(dir string) (bool, error) {
				return false, nil
			})
			errortest.AssertError(t, err, commonerrors.ErrTimeout)
		})
	}
}

func TestWalkUp_SymlinkTraversalIsLexical(t *testing.T) {
	printWarningOnWindows(t)
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "walkup-symlink")
			realStart := FilePathJoin(fs, root, "real", "parent", "start")
			require.NoError(t, fs.MkDirAll(realStart, 0o755))
			linkPath := FilePathJoin(fs, root, "link")
			err := fs.Symlink(FilePathJoin(fs, root, "real", "parent"), linkPath)
			skipIfLinksNotSupported(t, err)

			start := FilePathJoin(fs, linkPath, "start")
			var visited []string
			err = fs.WalkUp(context.Background(), start, func(dir string) (bool, error) {
				visited = append(visited, dir)
				return false, nil
			})
			require.NoError(t, err)
			assert.Contains(t, visited, linkPath)
			assert.NotContains(t, visited, FilePathJoin(fs, root, "real", "parent"))
		})
	}
}

func TestWalkUp_WindowsUNCTraversalHelper(t *testing.T) {
	if !platform.IsWindows() {
		t.Skip("windows-specific path handling")
	}
	fs := NewTestFilesystem(t, '\\')
	var visited []string
	err := walkUpLexically(context.Background(), fs, `\\server\share\dir\child`, nil, func(dir string) (bool, error) {
		visited = append(visited, dir)
		return false, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{`\\server\share\dir\child`, `\\server\share\dir`, `\\server\share\`}, visited)
}

func TestFindUp(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup")
			start := FilePathJoin(fs, root, "a", "b", "c")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			startPatternFile := FilePathJoin(fs, start, ".toolrc")
			writeWalkUpTestFile(t, fs, startPatternFile)
			ancestorPatternFile := FilePathJoin(fs, root, "a", "config.yaml")
			writeWalkUpTestFile(t, fs, ancestorPatternFile)
			farDir := FilePathJoin(fs, root, ".git")
			nearDir := FilePathJoin(fs, root, "a", "b", ".git")
			require.NoError(t, fs.MkDirAll(farDir, 0o755))
			require.NoError(t, fs.MkDirAll(nearDir, 0o755))
			configTree := FilePathJoin(fs, root, ".config", "tool")
			require.NoError(t, fs.MkDirAll(configTree, 0o755))
			nestedMatch := FilePathJoin(fs, configTree, "settings.json")
			writeWalkUpTestFile(t, fs, nestedMatch)

			found, err := fs.FindUp(context.Background(), start, ".toolrc")
			require.NoError(t, err)
			assert.Equal(t, startPatternFile, found)

			found, err = fs.FindUp(context.Background(), start, "*.yaml")
			require.NoError(t, err)
			assert.Equal(t, ancestorPatternFile, found)

			found, err = fs.FindUp(context.Background(), start, ".git")
			require.NoError(t, err)
			assert.Equal(t, nearDir, found)

			found, err = fs.FindUp(context.Background(), start, FilePathJoin(fs, ".config", "**", "*.json"))
			require.NoError(t, err)
			assert.Equal(t, nestedMatch, found)
		})
	}
}

func TestFindUp_NotFound(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup-missing")
			start := FilePathJoin(fs, root, "a", "b")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			_, err := fs.FindUp(context.Background(), start, "missing.*")
			errortest.AssertError(t, err, commonerrors.ErrNotFound)
		})
	}
}

func TestFindUp_InvalidPattern(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup-invalid")
			start := FilePathJoin(fs, root, "a")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			tests := []string{FilePathJoin(fs, root, "target"), FilePathJoin(fs, "..", "target"), FilePathJoin(fs, "child", "..", "..", "target"), "["}
			if runtime.GOOS == "windows" {
				tests = append(tests, `C:target`)
			}
			for _, pattern := range tests {
				_, err := fs.FindUp(context.Background(), start, pattern)
				errortest.AssertError(t, err, commonerrors.ErrInvalid)
			}
		})
	}
}

func TestFindUp_InvalidStart(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup-start")

			_, err := fs.FindUp(context.Background(), FilePathJoin(fs, root, "missing"), "target")
			errortest.AssertError(t, err, commonerrors.ErrNotFound)
		})
	}
}

func TestFindUp_ContextErrors(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("cancelled_for_fs_%v", fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup-cancel")
			start := FilePathJoin(fs, root, "a")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err := fs.FindUp(ctx, start, "target")
			errortest.AssertError(t, err, commonerrors.ErrCancelled)
		})

		t.Run(fmt.Sprintf("deadline_for_fs_%v", fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup-deadline")
			start := FilePathJoin(fs, root, "a")
			require.NoError(t, fs.MkDirAll(start, 0o755))

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			defer cancel()
			_, err := fs.FindUp(ctx, start, "target")
			errortest.AssertError(t, err, commonerrors.ErrTimeout)
		})
	}
}

func TestFindUp_PermissionErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test is not deterministic on Windows")
	}
	fs := NewStandardFileSystem()
	root := createWalkUpTestRoot(t, fs, "findup-perms")
	start := FilePathJoin(fs, root, "a", "b")
	require.NoError(t, fs.MkDirAll(start, 0o755))
	blocked := FilePathJoin(fs, root, "blocked")
	require.NoError(t, fs.MkDirAll(blocked, 0o755))
	require.NoError(t, fs.Chmod(blocked, 0o000))
	defer func() { _ = fs.Chmod(blocked, 0o755) }()

	_, err := fs.FindUp(context.Background(), start, FilePathJoin(fs, "blocked", "*"))
	assert.Error(t, err)
	assert.False(t, commonerrors.Any(err, commonerrors.ErrNotFound))
}

func TestFindUp_SymlinkTraversalIsLexical(t *testing.T) {
	printWarningOnWindows(t)
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			root := createWalkUpTestRoot(t, fs, "findup-symlink")
			realStart := FilePathJoin(fs, root, "real", "parent", "start")
			require.NoError(t, fs.MkDirAll(realStart, 0o755))
			resolvedOnlyMatch := FilePathJoin(fs, root, "real", "marker")
			writeWalkUpTestFile(t, fs, resolvedOnlyMatch)
			linkPath := FilePathJoin(fs, root, "link")
			err := fs.Symlink(FilePathJoin(fs, root, "real", "parent"), linkPath)
			skipIfLinksNotSupported(t, err)

			_, err = fs.FindUp(context.Background(), FilePathJoin(fs, linkPath, "start"), "marker")
			errortest.AssertError(t, err, commonerrors.ErrNotFound)
		})
	}
}

func TestValidateFindUpPattern(t *testing.T) {
	fs := NewStandardFileSystem()
	errortest.AssertError(t, validateFindUpPattern(fs, ""), commonerrors.ErrUndefined)
	errortest.AssertError(t, validateFindUpPattern(fs, FilePathJoin(fs, "..", "file")), commonerrors.ErrInvalid)
	require.NoError(t, validateFindUpPattern(fs, FilePathJoin(fs, "config", "**", "*.json")))
	if platform.IsWindows() {
		errortest.AssertError(t, validateFindUpPattern(fs, `C:file`), commonerrors.ErrInvalid)
	}
}

func createWalkUpTestRoot(t *testing.T, fs FS, prefix string) string {
	t.Helper()
	root := FilePathJoin(fs, t.TempDir(), prefix)
	require.NoError(t, fs.MkDirAll(root, 0o755))
	return root
}

func writeWalkUpTestFile(t *testing.T, fs FS, path string) {
	t.Helper()
	require.NoError(t, fs.MkDirAll(FilePathDir(fs, path), 0o755))
	require.NoError(t, fs.WriteFile(path, []byte("x"), 0o600))
}

func expectedWalkUpOrder(t *testing.T, fs FS, start string) []string {
	t.Helper()
	currentDir, err := fs.CurrentDirectory()
	require.NoError(t, err)
	current := FilePathClean(fs, FilePathAbs(fs, start, currentDir))
	var expected []string
	for {
		expected = append(expected, current)
		parent := FilePathDir(fs, current)
		if parent == current {
			return expected
		}
		current = parent
	}
}

func sanitiseTestName(name string) string {
	replacer := strings.NewReplacer("/", "-", `\\`, "-", " ", "-")
	return replacer.Replace(name)
}
