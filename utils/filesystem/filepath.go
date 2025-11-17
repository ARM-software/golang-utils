package filesystem

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// FilepathStem returns  the final path component, without its suffix. It's similar to `stem` in python's [pathlib](https://docs.python.org/3/library/pathlib.html#pathlib.PurePath.stem)
func FilepathStem(fp string) string {
	return FilePathStemOnFilesystem(GetGlobalFileSystem(), fp)
}

// FilepathParents returns a list of to the logical ancestors of the path, and it's similar to `parents` in python's [pathlib](https://docs.python.org/3/library/pathlib.html#pathlib.PurePath.parents)
func FilepathParents(fp string) []string {
	return FilePathParentsOnFilesystem(GetGlobalFileSystem(), fp)
}

// FilePathParentsOnFilesystem is similar to FilepathParents but with the ability to be applied to a particular filesystem.
func FilePathParentsOnFilesystem(fs FS, fp string) (parents []string) {
	cleanFp := FilePathClean(fs, fp)
	elements := strings.Split(cleanFp, string(fs.PathSeparator()))
	if len(elements) <= 1 {
		return
	}
	path := elements[0]
	if path == "" {
		elements = elements[1:]
		if len(elements) <= 1 {
			return
		}
		path = elements[0]
	}
	parents = append(parents, path)
	for i := 1; i < len(elements)-1; i++ {
		path = strings.Join([]string{path, elements[i]}, string(fs.PathSeparator()))
		parents = append(parents, path)
	}
	return
}

// FilePathJoin joins any number of path elements into a single path,
// separating them with the filesystem path separator.
// Its behaviour is similar to filepath.Join
func FilePathJoin(fs FS, element ...string) string {
	if fs == nil {
		return ""
	}

	if isSlashSeparator(fs) {
		return path.Join(element...)
	}

	elements := make([]string, len(element))
	for i := range elements {
		elements[i] = FilePathToPlatformPathSeparator(fs, element[i])
	}

	return FilePathFromPlatformPathSeparator(fs, filepath.Join(elements...))
}

// FilePathBase has the same behaviour as filepath.Base, but can handle different filesystems.
func FilePathBase(fs FS, fp string) string {
	if fs == nil {
		return ""
	}

	if isSlashSeparator(fs) {
		return path.Base(fp)
	}

	return FilePathFromPlatformPathSeparator(fs, filepath.Base(FilePathToPlatformPathSeparator(fs, fp)))
}

// FilePathDir has the same behaviour as filepath.Dir, but can handle different filesystems.
func FilePathDir(fs FS, fp string) string {
	if fs == nil {
		return ""
	}

	if isSlashSeparator(fs) {
		return path.Dir(fp)
	}
	return FilePathFromPlatformPathSeparator(fs, filepath.Dir(FilePathToPlatformPathSeparator(fs, fp)))
}

// FilePathIsAbs has the same behaviour as filepath.IsAbs, but can handle different filesystems.
func FilePathIsAbs(fs FS, fp string) bool {
	if fs == nil {
		return false
	}
	if isSlashSeparator(fs) {
		return path.IsAbs(fp)
	}
	return filepath.IsAbs(FilePathToPlatformPathSeparator(fs, fp))
}

// FilePathRel has the same behaviour as filepath.FilePathRel, but can handle different filesystems.
func FilePathRel(fs FS, basepath, targpath string) (rel string, err error) {
	if fs == nil {
		err = commonerrors.UndefinedVariable("filesystem specification")
		return
	}
	rel, err = filepath.Rel(FilePathToPlatformPathSeparator(fs, basepath), FilePathToPlatformPathSeparator(fs, targpath))
	if err != nil {
		return
	}
	rel = FilePathFromPlatformPathSeparator(fs, rel)
	return
}

// FilePathSplit has the same behaviour as filepath.Split, but can handle different filesystems
func FilePathSplit(fs FS, fp string) (dir, file string) {
	if fs == nil {
		return
	}

	if isSlashSeparator(fs) {
		return path.Split(fp)
	}
	dir, file = filepath.Split(FilePathToPlatformPathSeparator(fs, fp))
	dir = FilePathFromPlatformPathSeparator(fs, dir)
	return
}

// FilePathAbs tries to be similar to filepath.Abs behaviour but without using platform information or location.
func FilePathAbs(fs FS, fp, currentDirectory string) string {
	if FilePathIsAbs(fs, fp) {
		return FilePathClean(fs, fp)
	}
	return FilePathJoin(fs, currentDirectory, fp)
}

// FilePathToPlatformPathSeparator returns the result of replacing each separator character in path with a platform path separator character
func FilePathToPlatformPathSeparator(fs FS, path string) string {
	if fs.PathSeparator() == platform.PathSeparator {
		return path
	}
	return strings.ReplaceAll(path, string(fs.PathSeparator()), string(platform.PathSeparator))
}

// FilePathFromPlatformPathSeparator returns the result of replacing each platform path separator character in path with filesystem's path separator character
func FilePathFromPlatformPathSeparator(fs FS, path string) string {
	if fs.PathSeparator() == platform.PathSeparator {
		return path
	}
	return strings.ReplaceAll(path, string(platform.PathSeparator), string(fs.PathSeparator()))
}

// FilePathToSlash is filepath.ToSlash  but using filesystem path separator rather than platform's
func FilePathToSlash(fs FS, path string) string {
	if isSlashSeparator(fs) {
		return path
	}
	return strings.ReplaceAll(path, string(fs.PathSeparator()), "/")
}

// FilePathsToSlash just applying FilePathToSlash to a slice of path
func FilePathsToSlash(fs FS, path ...string) (toSlashes []string) {
	if len(path) == 0 {
		return
	}
	toSlashes = make([]string, len(path))
	for i := range path {
		toSlashes[i] = FilePathToSlash(fs, path[i])
	}
	return
}

// FilePathFromSlash is filepath.FromSlash but using filesystem path separator rather than platform's
func FilePathFromSlash(fs FS, path string) string {
	if isSlashSeparator(fs) {
		return path
	}
	return strings.ReplaceAll(path, "/", string(fs.PathSeparator()))
}

// FilePathsFromSlash just applying FilePathFromSlash to a slice of path
func FilePathsFromSlash(fs FS, path ...string) (fromlashes []string) {
	if len(path) == 0 {
		return
	}
	fromlashes = make([]string, len(path))
	for i := range path {
		fromlashes[i] = FilePathFromSlash(fs, path[i])
	}
	return
}

// FilePathStemOnFilesystem has the same behaviour as FilePathStem, but can handle different filesystems.
func FilePathStemOnFilesystem(fs FS, fp string) string {
	if fs == nil {
		return ""
	}
	return strings.TrimSuffix(FilePathBase(fs, fp), FilePathExt(fs, fp))
}

// FilePathExt has the same behaviour as filepath.Ext, but can handle different filesystems.
func FilePathExt(fs FS, fp string) string {
	if fs == nil {
		return ""
	}
	if isSlashSeparator(fs) {
		return path.Ext(fp)
	}
	return filepath.Ext(FilePathToPlatformPathSeparator(fs, fp))
}

// FilePathClean has the same behaviour as filepath.Clean, but can handle different filesystems.
func FilePathClean(fs FS, fp string) string {
	if fs == nil {
		return ""
	}

	if isSlashSeparator(fs) {
		return path.Clean(fp)
	}
	return FilePathFromPlatformPathSeparator(fs, filepath.Clean(FilePathToPlatformPathSeparator(fs, fp)))
}

// FilePathVolumeName has the same behaviour as filepath.VolumeName, but can handle different filesystems.
func FilePathVolumeName(fs FS, fp string) string {
	if fs == nil {
		return ""
	}

	return filepath.VolumeName(FilePathToPlatformPathSeparator(fs, fp))
}

func isSlashSeparator(fs FS) bool {
	if fs == nil {
		return false
	}
	return fs.PathSeparator() == '/'
}

// FileTreeDepth returns the depth of a file in a tree starting from root
func FileTreeDepth(fs FS, root, filePath string) (depth int64, err error) {
	if reflection.IsEmpty(filePath) {
		return
	}
	rel, err := fs.ConvertToRelativePath(root, filePath)
	if err != nil {
		return
	}
	diff := rel[0]
	if reflection.IsEmpty(diff) {
		return
	}
	diff = FilePathToSlash(fs, diff)
	depth = int64(len(strings.Split(diff, "/")) - 1)
	return
}

// EvalSymlinks has the same behaviour as  filepath.EvalSymlinks , but can handle different filesystems.
func EvalSymlinks(fs FS, pathWithSymlinks string) (populatedPath string, err error) {
	if fs == nil {
		return "", commonerrors.UndefinedVariable("filesystem")
	}

	// FIXME the following is only true for osfs
	// Use https://github.com/spf13/afero/issues/562 whenever it is made available.
	p, err := filepath.EvalSymlinks(FilePathToPlatformPathSeparator(fs, pathWithSymlinks))
	if err != nil {
		err = commonerrors.WrapIfNotCommonErrorf(commonerrors.ErrUnexpected, ConvertFileSystemError(err), "could not evaluate the path '%v'", pathWithSymlinks)
		return
	}

	populatedPath = FilePathFromPlatformPathSeparator(fs, p)
	return
}

// EndsWithPathSeparator states whether a path is ending with a path separator of not
func EndsWithPathSeparator(fs FS, filePath string) bool {
	return strings.HasSuffix(filePath, "/") || strings.HasSuffix(filePath, string(fs.PathSeparator()))
}

// NewPathValidationRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid not empty path.
// `when` describes whether the rule is enforced or not
func NewPathValidationRule(filesystem FS, when bool) validation.Rule {
	return &pathValidationRule{condition: when, filesystem: filesystem}
}

// NewOSPathValidationRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid path for the Operating System's filesystem.
// `when` describes whether the rule is enforced or not
func NewOSPathValidationRule(when bool) validation.Rule {
	return NewPathValidationRule(GetGlobalFileSystem(), when)
}

type pathValidationRule struct {
	condition  bool
	filesystem FS
}

func (r *pathValidationRule) Validate(value interface{}) error {
	err := validation.Required.When(r.condition).Validate(value)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrUndefined, err, "path [%v] is required", value)
	}
	if !r.condition {
		return nil
	}
	pathString, err := validation.EnsureString(value)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "path [%v] must be a string", value)
	}
	pathString = strings.TrimSpace(pathString)
	// This check is here because it validates the path on any platform (it is a cross-platform check)
	// Indeed if the path exists, then it can only be valid.
	if r.filesystem.Exists(pathString) {
		return nil
	}

	// Inspired from https://github.com/go-playground/validator/blob/84254aeb5a59e615ec0b66ab53b988bc0677f55e/baked_in.go#L1604 and https://stackoverflow.com/questions/35231846/golang-check-if-string-is-valid-path
	if pathString == "" {
		return commonerrors.Newf(commonerrors.ErrUndefined, "the path [%v] is empty", value)
	}
	// This check is to catch errors on Linux. It does not work as well on Windows.
	if _, err := r.filesystem.Stat(pathString); err != nil {
		switch t := err.(type) {
		case *fs.PathError:
			if t.Err == syscall.EINVAL {
				return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "the path [%v] has invalid characters", value)
			}
		default:
			// make the linter happy
		}
	}
	// The following case is not caught on Windows by the check above.
	if strings.Contains(pathString, "\n") {
		return commonerrors.Newf(commonerrors.ErrInvalid, "the path [%v] has carriage returns characters", value)
	}

	// TODO add platform validation checks: e.g. https://learn.microsoft.com/en-gb/windows/win32/fileio/naming-a-file?redirectedfrom=MSDN on windows

	return nil
}

// NewPathExistRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid not empty path and actually exists.
// `when` describes whether the rule is enforced or not.
func NewPathExistRule(filesystem FS, when bool) validation.Rule {
	return &pathExistValidationRule{filesystem: filesystem, condition: when}
}

// NewOSPathExistRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid path for the Operating system's filesystem and actually exists.
// `when` describes whether the rule is enforced or not.
func NewOSPathExistRule(when bool) validation.Rule {
	return NewPathExistRule(GetGlobalFileSystem(), when)
}

type pathExistValidationRule struct {
	condition  bool
	filesystem FS
}

func (r *pathExistValidationRule) Validate(value interface{}) error {
	err := NewPathValidationRule(r.filesystem, r.condition).Validate(value)
	if err != nil {
		return err
	}
	if !r.condition {
		return nil
	}
	path, err := validation.EnsureString(value)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "path [%v] must be a string", value)
	}
	if !r.filesystem.Exists(path) {
		err = commonerrors.Newf(commonerrors.ErrNotFound, "path [%v] does not exist", path)
	}
	return err
}
