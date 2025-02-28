package filesystem

import (
	"embed"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func newEmbedFSAdapter(fs *embed.FS) (afero.Fs, error) {
	if fs == nil {
		return nil, commonerrors.UndefinedVariable("embedded file system")
	}
	return afero.NewReadOnlyFs(afero.FromIOFS{
		FS: *fs,
	}), nil
}
