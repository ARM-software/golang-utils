package filecache

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// FsEntryRetriever implements EntrySource by copying files or directories from a source filesystem into a cache filesystem.
type FsEntryRetriever struct {
	fs            filesystem.FS
	basePath      string
	cachefs       filesystem.FS
	cacheBasePath string
}

// FetchEntry locates the resource identified by `key` within the provider’s source filesystem under `basePath`,
// copies it into `cacheDir` using `cacheFilesystem`, and returns the absolute path of the cached entry.
func (p *FsEntryRetriever) FetchEntry(ctx context.Context, key string) (string, error) {
	srcPath := filesystem.FilePathJoin(p.fs, p.basePath, key)
	if !p.fs.Exists(srcPath) {
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

func (p *FsEntryRetriever) SetCacheDir(cacheFs filesystem.FS, cacheDir string) error {
	if cacheFs == nil {
		return commonerrors.New(commonerrors.ErrUndefined, "the cache filesystem cannot be nil")
	}

	p.cachefs = cacheFs

	if !p.cachefs.Exists(cacheDir) {
		return commonerrors.Newf(commonerrors.ErrNotFound, "cannot access '%s', No such file or directory", cacheDir)
	}

	p.cacheBasePath = cacheDir

	return nil
}

func NewFsFileCache(ctx context.Context, srcFilesystem, cacheFilesystem filesystem.FS, basePath string, config *FileCacheConfig) (IFileCache, error) {
	fsProvider := &FsEntryRetriever{
		fs:       srcFilesystem,
		basePath: basePath,
	}

	if err := fsProvider.SetCacheDir(cacheFilesystem, config.CachePath); err != nil {
		return nil, err
	}

	return NewGenericFileCache(ctx, cacheFilesystem, fsProvider, config)
}
