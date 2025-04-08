package sharedcache

import (
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

//go:generate go run github.com/dmarkham/enumer -type=CacheType -text -json -yaml
type CacheType int

const (
	CacheMutable CacheType = iota
	CacheImmutable
)

var (
	CacheTypes = []CacheType{CacheMutable, CacheImmutable}
)

func NewCache(cacheType CacheType, fs filesystem.FS, cfg *Configuration) (ISharedCacheRepository, error) {
	switch cacheType {
	case CacheMutable:
		return NewSharedMutableCacheRepository(cfg, fs)
	case CacheImmutable:
		return NewSharedImmutableCacheRepository(cfg, fs)
	}
	return nil, fmt.Errorf("%w: unknown cache type [%v]", commonerrors.ErrNotFound, cacheType)
}
