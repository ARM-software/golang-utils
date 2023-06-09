package filesystem

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type embedFsAdapter struct {
	afero.Fs
}

func (e *embedFsAdapter) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	if flag != os.O_RDONLY {
		return nil, fmt.Errorf("%w: embed.FS is readonly", commonerrors.ErrUnsupported)
	}
	return e.Open(name)
}

func newEmbedFSAdapter(fs *embed.FS) (afero.Fs, error) {
	if fs == nil {
		return nil, fmt.Errorf("%w: missing filesystem", commonerrors.ErrUndefined)
	}
	return &embedFsAdapter{
		Fs: afero.FromIOFS{
			FS: *fs,
		},
	}, nil
}
