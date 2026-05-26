package filesystem

import (
	"io/fs"
	"strings"
	"syscall"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

type filesystemValidationRule struct {
	condition  bool
	filesystem FS
}

func (r *filesystemValidationRule) Validate(value any) error {
	err := validation.Required.When(r.condition).Validate(value)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrUndefined, err, "missing value")
	}
	if r.filesystem == nil {
		return commonerrors.UndefinedVariable("filesystem to consider")
	}
	return nil
}

type pathValidationRule struct {
	filesystemValidationRule
}

func (r *pathValidationRule) Validate(value any) error {
	err := r.filesystemValidationRule.Validate(value)
	if err != nil {
		return err
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

// NewPathValidationRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid not empty path.
// `when` describes whether the rule is enforced or not.
func NewPathValidationRule(filesystem FS, when bool) validation.Rule {
	return &pathValidationRule{
		filesystemValidationRule: filesystemValidationRule{
			condition:  when,
			filesystem: filesystem,
		},
	}
}

// NewOSPathValidationRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid path for the operating system's
// filesystem.
// `when` describes whether the rule is enforced or not.
func NewOSPathValidationRule(when bool) validation.Rule {
	return NewPathValidationRule(GetGlobalFileSystem(), when)
}

type pathExistValidationRule struct {
	pathValidationRule
}

func (r *pathExistValidationRule) Validate(value any) error {
	err := r.pathValidationRule.Validate(value)
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

// NewPathExistRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid not empty path and actually
// exists.
// `when` describes whether the rule is enforced or not.
func NewPathExistRule(filesystem FS, when bool) validation.Rule {
	return &pathExistValidationRule{
		pathValidationRule: pathValidationRule{
			filesystemValidationRule: filesystemValidationRule{
				condition:  when,
				filesystem: filesystem,
			},
		},
	}
}

// NewOSPathExistRule returns a validation rule to use in configuration.
// The rule checks whether a string is a valid path for the operating system's
// filesystem and actually exists.
// `when` describes whether the rule is enforced or not.
func NewOSPathExistRule(when bool) validation.Rule {
	return NewPathExistRule(GetGlobalFileSystem(), when)
}

type pathExtensionValidationRule struct {
	pathValidationRule
	extensions []string
}

// NewPathExtensionRule returns a validation rule that checks whether a path has
// an extension present in the supplied list.
// `when` describes whether the rule is enforced or not.
func NewPathExtensionRule(filesystem FS, when bool, extensions ...string) validation.Rule {
	return &pathExtensionValidationRule{
		pathValidationRule: pathValidationRule{
			filesystemValidationRule: filesystemValidationRule{
				condition:  when,
				filesystem: filesystem,
			},
		},
		extensions: normaliseExtensions(filesystem, extensions...),
	}

}

// NewOSPathExtensionRule returns a validation rule that checks whether a path
// on the global filesystem has an extension present in the supplied list.
// `when` describes whether the rule is enforced or not.
func NewOSPathExtensionRule(when bool, extensions ...string) validation.Rule {
	return NewPathExtensionRule(GetGlobalFileSystem(), when, extensions...)
}

func (r *pathExtensionValidationRule) Validate(value any) error {
	err := r.pathValidationRule.Validate(value)
	if err != nil {
		return err
	}
	if !r.condition {
		return nil
	}
	if len(r.extensions) == 0 {
		return commonerrors.UndefinedVariable("allowed file extensions")
	}

	pathString, err := validation.EnsureString(value)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "path [%v] must be a string", value)
	}

	extension := strings.ToLower(FilePathExt(r.filesystem, strings.TrimSpace(pathString)))
	if reflection.IsEmpty(extension) {
		return commonerrors.Newf(commonerrors.ErrNoExtension, "path [%v] has no extension", value)
	}
	if collection.In(r.extensions, extension, collection.StringMatch) {
		return nil
	}
	return commonerrors.Newf(commonerrors.ErrInvalid, "path [%v] must have one of the extensions %v", value, r.extensions)

}

func normaliseExtensions(fs FS, extensions ...string) []string {
	return collection.Map[string, string](extensions, func(extension string) string {
		extension = strings.TrimSpace(extension)
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		return strings.ToLower(FilePathClean(fs, extension))
	})
}
