package filesystem

import (
	"os"

	"github.com/spf13/afero"
)

type extendedFile struct {
	afero.File
	onCloseCallBack func() error
}

func (f *extendedFile) Read(p []byte) (n int, err error) {
	n, err = f.File.Read(p)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = f.File.ReadAt(p, off)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Seek(offset int64, whence int) (n int64, err error) {
	n, err = f.File.Seek(offset, whence)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Write(p []byte) (n int, err error) {
	n, err = f.File.Write(p)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = f.File.WriteAt(p, off)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Name() string {
	return f.File.Name()
}

func (f *extendedFile) Readdir(count int) (i []os.FileInfo, err error) {
	i, err = f.File.Readdir(count)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Readdirnames(n int) (names []string, err error) {
	names, err = f.File.Readdirnames(n)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Stat() (i os.FileInfo, err error) {
	i, err = f.File.Stat()
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Sync() error {
	return ConvertFileSystemError(f.File.Sync())
}

func (f *extendedFile) Truncate(size int64) error {
	return ConvertFileSystemError(f.File.Truncate(size))
}

func (f *extendedFile) WriteString(s string) (ret int, err error) {
	ret, err = f.File.WriteString(s)
	err = ConvertFileSystemError(err)
	return
}

func (f *extendedFile) Close() (err error) {
	err = f.File.Close()
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	if f.onCloseCallBack != nil {
		err = f.onCloseCallBack()
	}
	return
}

func (f *extendedFile) Fd() (fd uintptr) {
	if correctFile, ok := retrieveSubFile(f.File).(interface {
		Fd() uintptr
	}); ok {
		fd = correctFile.Fd()
	} else {
		fd = uintptr(UnsetFileHandle)
	}
	return
}

func convertFile(getFile func() (afero.File, error), onCloseCallBack func() error) (f File, err error) {
	file, err := getFile()
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	return convertToExtendedFile(file, onCloseCallBack)
}

func convertToExtendedFile(file afero.File, onCloseCallBack func() error) (File, error) {
	return &extendedFile{
		File: file,
		onCloseCallBack: func() error {
			return ConvertFileSystemError(onCloseCallBack())
		},
	}, nil
}

// ConvertToOSFile converts a file to a `os` implementation of a file for certain use-cases where functions have not moved to using `fs.File`.
func ConvertToOSFile(f File) (osFile *os.File) {
	if f == nil {
		return
	}
	osFile, ok := f.(*os.File)
	if ok {
		return
	}
	osFile = os.NewFile(f.Fd(), f.Name())
	return
}
