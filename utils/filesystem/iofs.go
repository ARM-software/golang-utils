package filesystem

import (
	"io/fs"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type wrappedFS struct {
	filesystem FS
}

func (w *wrappedFS) Open(name string) (fs.File, error) {
	return w.filesystem.GenericOpen(name)
}

// ConvertToIOFilesystem converts a filesystem FS to a io/fs FS
func ConvertToIOFilesystem(filesystem FS) (fs.FS, error) {
	if filesystem == nil {
		return nil, commonerrors.New(commonerrors.ErrUndefined, "missing filesystem")
	}
	return &wrappedFS{filesystem: filesystem}, nil
}

// ConvertFromIOFilesystem converts an io/fs FS into a FS
func ConvertFromIOFilesystem(filesystem fs.FS) (FS, error) {
	if filesystem == nil {
		return nil, commonerrors.New(commonerrors.ErrUndefined, "missing filesystem")
	}
	return NewVirtualFileSystem(afero.FromIOFS{FS: filesystem}, Custom, IdentityPathConverterFunc), nil
}
