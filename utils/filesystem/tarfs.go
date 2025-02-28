package filesystem

import (
	"archive/tar"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func newTarFSAdapterFromReader(reader *tar.Reader) (afero.Fs, error) {
	if reader == nil {
		return nil, commonerrors.New(commonerrors.ErrUndefined, "missing reader")
	}
	return afero.NewReadOnlyFs(tarfs.New(reader)), nil
}

func newTarFSAdapterFromFilePath(fs FS, tarFilePath string, limits ILimits) (tarFs afero.Fs, tarFile File, err error) {
	if fs == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing filesystem to use for finding the tar file")
		return
	}
	tarReader, tarFile, err := newTarReader(fs, tarFilePath, limits, 0)
	if err != nil && tarFile != nil {
		subErr := tarFile.Close()
		if subErr == nil {
			tarFile = nil
		}
		return
	}
	tarFs, err = newTarFSAdapterFromReader(tarReader)
	return
}
