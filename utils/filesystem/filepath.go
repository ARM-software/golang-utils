package filesystem

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// FilepathStem returns  the final path component, without its suffix. It's similar to `stem` in python's [pathlib](https://docs.python.org/3/library/pathlib.html#pathlib.PurePath.stem)
func FilepathStem(fp string) string {
	return strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp))
}

// FilepathParents returns a list of to the logical ancestors of the path and it's similar to `parents` in python's [pathlib](https://docs.python.org/3/library/pathlib.html#pathlib.PurePath.parents)
func FilepathParents(fp string) []string {
	return FilePathParentsOnFilesystem(GetGlobalFileSystem(), fp)
}

// FilePathParentsOnFilesystem is similar to FilepathParents but with the ability to be applied to a particular filesystem.
func FilePathParentsOnFilesystem(fs FS, fp string) (parents []string) {
	var cleanFp string
	if fs.GetType() == Embed {
		cleanFp = path.Clean(fp)
	} else {
		cleanFp = filepath.Clean(fp)
	}
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

	if fs.GetType() == Embed {
		return path.Join(element...)
	}

	return filepath.Join(element...)
}

// Same behaviour as filepath.Base, but can handle different filesystems.
func Base(fs FS, fp string) string {
	if fs == nil {
		return ""
	}

	if fs.GetType() == Embed {
		return path.Base(fp)
	}

	return filepath.Base(fp)
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
	diff = strings.ReplaceAll(diff, string(fs.PathSeparator()), "/")
	depth = int64(len(strings.Split(diff, "/")) - 1)
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
