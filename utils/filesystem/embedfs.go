package filesystem

import (
	"embed"
	"fmt"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func newEmbedFSAdapter(fs *embed.FS) (afero.Fs, error) {
	if fs == nil {
		return nil, fmt.Errorf("%w: missing filesystem", commonerrors.ErrUndefined)
	}
	return afero.NewReadOnlyFs(afero.FromIOFS{
		FS: *fs,
	}), nil
}
