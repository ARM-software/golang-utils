package filesystem

import (
	"archive/tar"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func newTarReader(fs FS, source string, limits ILimits, currentDepth int64) (tarReader *tar.Reader, file File, err error) {
	if fs == nil {
		err = fmt.Errorf("%w: missing file system", commonerrors.ErrUndefined)
		return
	}
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}
	if limits.Apply() && limits.GetMaxDepth() >= 0 && currentDepth > limits.GetMaxDepth() {
		err = fmt.Errorf("%w: depth [%v] of tar file [%v] is beyond allowed limits (max: %v)", commonerrors.ErrTooLarge, currentDepth, source, limits.GetMaxDepth())
		return
	}

	if !fs.Exists(source) {
		err = fmt.Errorf("%w: could not find archive [%v]", commonerrors.ErrNotFound, source)
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
		err = fmt.Errorf("%w: tar file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, source, tarFileSize, limits.GetMaxFileSize())
		return
	}

	tarReader = tar.NewReader(file)

	return
}
