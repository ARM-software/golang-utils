package filecache

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

type FsEntryProvider struct {
	fs       filesystem.FS
	basePath string
}

func (p *FsEntryProvider) StoreEntry(ctx context.Context, key string, cacheFilesystem filesystem.FS, cacheDir string) (string, error) {
	srcPath := filesystem.FilePathJoin(p.fs, p.basePath, key)
	if _, statErr := filesystem.Stat(srcPath); statErr != nil {
		return "", statErr
	}

	destPath := filesystem.FilePathJoin(p.fs, cacheDir, key)
	cpErr := filesystem.CopyBetweenFS(ctx, p.fs, srcPath, cacheFilesystem, destPath)
	if cpErr != nil {
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
