package filesystem

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func printWarningOnWindows(t *testing.T) {
	t.Helper()
	if platform.IsWindows() {
		t.Log("⚠️ In order to run TestLink on Windows, Developer mode must be enabled: https://github.com/golang/go/pull/24307")
	}
}

func skipIfLinksNotSupported(t *testing.T, err error) {
	t.Helper()
	if commonerrors.Any(err, commonerrors.ErrNotImplemented, commonerrors.ErrForbidden, commonerrors.ErrUnsupported) {
		t.Skipf("⚠️ links not supported on this system: %v", err)
	} else {
		require.NoError(t, err)
	}
}

func TestLink(t *testing.T) {
	printWarningOnWindows(t)
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-link-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := fmt.Sprintf("This is a test sentence!!! %v", faker.Sentence())
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)
			err = fs.WriteFile(tmpFile, []byte(txt), 0755)
			require.NoError(t, err)

			symlink := FilePathJoin(fs, tmpDir, "symlink-tofile")
			hardlink := FilePathJoin(fs, tmpDir, "hardlink-tofile")

			err = fs.Symlink(tmpFile, symlink)
			skipIfLinksNotSupported(t, err)

			err = fs.Link(tmpFile, hardlink)
			require.NoError(t, err)

			assert.True(t, fs.Exists(symlink))
			assert.True(t, fs.Exists(hardlink))

			isLink, err := fs.IsLink(symlink)
			require.NoError(t, err)
			assert.True(t, isLink)

			isFile, err := fs.IsFile(symlink)
			require.NoError(t, err)
			assert.True(t, isFile)

			isLink, err = fs.IsLink(hardlink)
			require.NoError(t, err)
			assert.False(t, isLink)

			isFile, err = fs.IsFile(hardlink)
			require.NoError(t, err)
			assert.True(t, isFile)

			link, err := fs.Readlink(symlink)
			require.NoError(t, err)
			assert.Equal(t, tmpFile, link)

			link, err = fs.Readlink(hardlink)
			require.Error(t, err)
			assert.Empty(t, link)

			bytes, err := fs.ReadFile(symlink)
			require.NoError(t, err)
			assert.Equal(t, txt, string(bytes))

			bytes, err = fs.ReadFile(hardlink)
			require.NoError(t, err)
			assert.Equal(t, txt, string(bytes))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestEvalSymlinks_ResolvesToRealPath(t *testing.T) {
	printWarningOnWindows(t)
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-eval-link-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()
			realTmpDir, err := fs.TempDir(tmpDir, "real-dir-")
			require.NoError(t, err)
			symTmpDir, err := fs.TempDir(tmpDir, "sym-dir-")
			require.NoError(t, err)
			// real target: <realTmpDir>/real/nested/file.txt
			realDir := FilePathJoin(fs, realTmpDir, "real", "nested")
			require.NoError(t, fs.MkDir(realDir))
			realFile := FilePathJoin(fs, realDir, "file.txt")
			expectedContent := fmt.Sprintf("hello world! hello %v! %v", faker.Name(), faker.Sentence())
			require.NoError(t, fs.WriteFile(realFile, []byte(expectedContent), 0o644))
			currentDir, err := fs.CurrentDirectory()
			require.NoError(t, err)
			expectedAbs := FilePathAbs(fs, realFile, currentDir)

			// First symlink: <symTmpDir>/random -> <realTmpDir>/real/nested
			symDir := FilePathJoin(fs, symTmpDir, faker.Word())
			err = fs.Symlink(realDir, symDir)
			skipIfLinksNotSupported(t, err)

			// symlink to a symlink
			symTmpDir2, err := fs.TempDir(tmpDir, "sym-dir2-")
			require.NoError(t, err)
			symAgain := FilePathJoin(fs, symTmpDir2, fmt.Sprintf("%v-sym2sym", faker.Word()))
			err = fs.Symlink(symDir, symAgain)
			skipIfLinksNotSupported(t, err)

			pathThroughSymlinks := FilePathJoin(fs, symAgain, "file.txt")
			symlinkFile := FilePathJoin(fs, symAgain, "symfile.txt")
			err = fs.Symlink(pathThroughSymlinks, symlinkFile)
			skipIfLinksNotSupported(t, err)

			resolved, err := EvalSymlinks(fs, pathThroughSymlinks)
			require.NoError(t, err)

			resolvedAbs := FilePathAbs(fs, resolved, currentDir)
			resolvedAbs = FilePathClean(fs, resolvedAbs)
			expectedAbs = FilePathClean(fs, expectedAbs)

			resolved2, err := EvalSymlinks(fs, symlinkFile)
			require.NoError(t, err)
			resolvedAbs2 := FilePathAbs(fs, resolved2, currentDir)
			resolvedAbs2 = FilePathClean(fs, resolvedAbs2)

			if platform.IsWindows() {
				assert.True(t, strings.EqualFold(resolvedAbs, expectedAbs))
				assert.True(t, strings.EqualFold(resolvedAbs2, expectedAbs))
			} else {
				assert.Equal(t, expectedAbs, resolvedAbs)
				assert.Equal(t, expectedAbs, resolvedAbs2)
			}

			content, err := fs.ReadFile(symlinkFile)
			require.NoError(t, err)
			assert.Equal(t, string(content), expectedContent)

		})
	}
}

func TestEvalSymlinks_notExist(t *testing.T) {
	printWarningOnWindows(t)
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-eval-link-not-exist-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()
			_, err = EvalSymlinks(fs, "notexist")
			errortest.AssertError(t, err, commonerrors.ErrNotFound)
			dest := FilePathJoin(fs, tmpDir, "link")
			err = fs.Symlink("notexist", dest)
			skipIfLinksNotSupported(t, err)
			_, err = EvalSymlinks(fs, dest)
			errortest.AssertError(t, err, commonerrors.ErrNotFound)
		})
	}
}
