package simplecache

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

type ISimpleCache interface {
	Add(context.Context, string, filesystem.FS, string) error
	Remove(context.Context, string) error
	Contains(context.Context, string) (bool, error)
	Restore(context.Context, string) error
	Close(context.Context) error
}
