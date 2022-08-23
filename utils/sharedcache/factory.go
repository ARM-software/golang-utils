package sharedcache

import (
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

const (
	CacheMutable int = iota
	CacheImmutable
)

var (
	CacheTypes = []int{CacheMutable, CacheImmutable}
)

func NewCache(cacheType int, fs filesystem.FS, cfg *Configuration) (ISharedCacheRepository, error) {
	switch cacheType {
	case CacheMutable:
		return NewSharedMutableCacheRepository(cfg, fs)
	case CacheImmutable:
		return NewSharedImmutableCacheRepository(cfg, fs)
	}
	return nil, fmt.Errorf("%w: unknown cache type [%v]", commonerrors.ErrNotFound, cacheType)
}
