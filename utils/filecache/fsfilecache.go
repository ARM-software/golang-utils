package filecache

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// FsEntryProvider implements EntrySource by copying files or directories from a source filesystem into a cache filesystem.
type FsEntryProvider struct {
	fs            filesystem.FS
	cachefs       filesystem.FS
	basePath      string
	cacheBasePath string
}

// FetchEntry locates the resource identified by `key` within the provider’s source filesystem under `basePath`,
// copies it into `cacheDir` using `cacheFilesystem`, and returns the absolute path of the cached entry.
func (p *FsEntryProvider) FetchEntry(ctx context.Context, key string) (string, error) {
	srcPath := filesystem.FilePathJoin(p.fs, p.basePath, key)
	if exists := p.fs.Exists(srcPath); !exists {
		return "", commonerrors.Newf(commonerrors.ErrNotFound, "cannot access '%s', No such file or directory", srcPath)
	}

	tmpKey := fmt.Sprintf("%v-tmp", key)
	destPathTmp := filesystem.FilePathJoin(p.fs, p.cacheBasePath, tmpKey)
	cpErr := filesystem.CopyBetweenFS(ctx, p.fs, srcPath, p.cachefs, destPathTmp)
	if cpErr != nil {
		return "", cpErr
	}

	destPath := filesystem.FilePathJoin(p.fs, p.cacheBasePath, key)
	mvErr := p.cachefs.Move(destPathTmp, destPath)
	if mvErr != nil {
		return "", cpErr
	}

	return destPath, nil
}

func (p *FsEntryProvider) SetCacheFilesystem(fs filesystem.FS) error {
	if fs == nil {
		return commonerrors.New(commonerrors.ErrInvalid, "the cache filesystem cannot be nil")
	}

	p.cachefs = fs

	return nil
}

func (p *FsEntryProvider) SetCacheDir(dir string) error {
	if exists := p.cachefs.Exists(dir); !exists {
		return commonerrors.Newf(commonerrors.ErrNotFound, "cannot access '%s', No such file or directory", dir)
	}

	p.cacheBasePath = dir

	return nil
}

func NewFsFileCache(ctx context.Context, srcFilesystem, cacheFilesystem filesystem.FS, basePath string, config *FileCacheConfig) (IFileCache, error) {
	fsProvider := &FsEntryProvider{
		fs:       srcFilesystem,
		basePath: basePath,
	}

	if err := fsProvider.SetCacheFilesystem(cacheFilesystem); err != nil {
		return nil, err
	}

	if err := fsProvider.SetCacheDir(config.CachePath); err != nil {
		return nil, err
	}

	return NewGenericFileCache(ctx, cacheFilesystem, fsProvider, config)
}
