package filecache

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// FsEntryProvider implements EntrySource by copying files or directories from a source filesystem into a cache filesystem.
type FsEntryProvider struct {
	fs       filesystem.FS
	basePath string
}

// FetchEntry locates the resource identified by `key` within the provider’s source filesystem under `basePath`,
// copies it into `cacheDir` using `cacheFilesystem`, and returns the absolute path of the cached entry.
func (p *FsEntryProvider) FetchEntry(ctx context.Context, key string, cacheFilesystem filesystem.FS, cacheDir string) (string, error) {
	srcPath := filesystem.FilePathJoin(p.fs, p.basePath, key)
	if exists := p.fs.Exists(srcPath); !exists {
		return "", commonerrors.Newf(commonerrors.ErrNotFound, "cannot access '%s', No such file or directory", srcPath)
	}

	tmpKey := fmt.Sprintf("%v-tmp", key)
	destPathTmp := filesystem.FilePathJoin(p.fs, cacheDir, tmpKey)
	cpErr := filesystem.CopyBetweenFS(ctx, p.fs, srcPath, cacheFilesystem, destPathTmp)
	if cpErr != nil {
		return "", cpErr
	}

	destPath := filesystem.FilePathJoin(p.fs, cacheDir, key)
	mvErr := cacheFilesystem.Move(destPathTmp, destPath)
	if mvErr != nil {
		return "", cpErr
	}

	return destPath, nil
}

func NewFsFileCache(ctx context.Context, srcFilesystem, cacheFilesystem filesystem.FS, basePath string, config *FileCacheConfig) (IFileCache, error) {
	fsProvider := &FsEntryProvider{
		fs:       srcFilesystem,
		basePath: basePath,
	}

	return NewGenericFileCache(ctx, cacheFilesystem, fsProvider, config)
}
