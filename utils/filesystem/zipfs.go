package filesystem

import (
	"archive/zip"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func newZipFSAdapterFromReader(reader *zip.Reader) (afero.Fs, error) {
	if reader == nil {
		return nil, fmt.Errorf("%w: missing reader", commonerrors.ErrUndefined)
	}
	return afero.NewReadOnlyFs(zipfs.New(reader)), nil
}

func newZipFSAdapterFromFilePath(fs FS, zipFilePath string, limits ILimits) (zipFs afero.Fs, zipFile File, err error) {
	if fs == nil {
		err = fmt.Errorf("%w: missing filesystem to use for finding the zip file", commonerrors.ErrUndefined)
		return
	}
	zipReader, zipFile, err := newZipReader(fs, zipFilePath, limits, 0)
	if err != nil && zipFile != nil {
		err = ConvertFileSystemError(err)
		subErr := zipFile.Close()
		if subErr == nil {
			zipFile = nil
		}
		return
	}
	zipFs, err = newZipFSAdapterFromReader(zipReader)
	return
}
