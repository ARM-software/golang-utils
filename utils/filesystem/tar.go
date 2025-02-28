package filesystem

import (
	"archive/tar"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func newTarReader(fs FS, source string, limits ILimits, currentDepth int64) (tarReader *tar.Reader, file File, err error) {
	if fs == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing file system")
		return
	}
	if limits == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing file system limits")
		return
	}
	if limits.Apply() && limits.GetMaxDepth() >= 0 && currentDepth > limits.GetMaxDepth() {
		err = commonerrors.Newf(commonerrors.ErrTooLarge, "depth [%v] of tar file [%v] is beyond allowed limits (max: %v)", currentDepth, source, limits.GetMaxDepth())
		return
	}

	if !fs.Exists(source) {
		err = commonerrors.Newf(commonerrors.ErrNotFound, "could not find archive [%v]", source)
		return
	}

	info, err := fs.Lstat(source)
	if err != nil {
		return
	}
	file, err = fs.GenericOpen(source)
	if err != nil {
		return
	}

	tarFileSize := info.Size()

	if limits.Apply() && tarFileSize > limits.GetMaxFileSize() {
		err = commonerrors.Newf(commonerrors.ErrTooLarge, "tar file [%v] is too big (%v B) and beyond limits (max: %v B)", source, tarFileSize, limits.GetMaxFileSize())
		return
	}

	tarReader = tar.NewReader(file)

	return
}
