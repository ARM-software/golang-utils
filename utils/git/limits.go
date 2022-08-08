package git

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/config"
)

type noLimits struct {
}

func (n *noLimits) Apply() bool {
	return false
}

func (n *noLimits) GetMaxFileSize() int64 {
	return 0
}

func (n *noLimits) GetMaxTotalSize() int64 {
	return 0
}

func (n *noLimits) GetMaxFileCount() int64 {
	return 0
}

func (n *noLimits) GetMaxTreeDepth() int64 {
	return 0
}

func (n *noLimits) GetMaxEntries() int64 {
	return 0
}

func (n *noLimits) Validate() error {
	return nil
}

// Limits defines file system limits
type Limits struct {
	MaxFileSize  int64 `mapstructure:"max_file_size"`
	MaxTotalSize int64 `mapstructure:"max_total_size"`
	MaxFileCount int64 `mapstructure:"max_file_count"`
	MaxTreeDepth int64 `mapstructure:"max_tree_depth"`
	MaxEntries   int64 `mapstructure:"max_entries"`
}

func (l *Limits) Apply() bool {
	return true
}

func (l *Limits) GetMaxFileSize() int64 {
	return l.MaxFileSize
}

func (l *Limits) GetMaxTotalSize() int64 {
	return l.MaxTotalSize
}

func (l *Limits) GetMaxFileCount() int64 {
	return l.MaxFileCount
}

func (l *Limits) GetMaxTreeDepth() int64 {
	return l.MaxTreeDepth
}
func (l *Limits) GetMaxEntries() int64 {
	return l.MaxEntries
}

func (l *Limits) Validate() error {
	validation.ErrorTag = "mapstructure"

	// Validate Embedded Structs
	err := config.ValidateEmbedded(l)
	if err != nil {
		return err
	}
	return validation.ValidateStruct(l,
		validation.Field(&l.MaxFileSize, validation.Required.When(l.Apply())),
		validation.Field(&l.MaxTotalSize, validation.Required.When(l.Apply())),
		validation.Field(&l.MaxFileCount, validation.Required.When(l.Apply())),
		validation.Field(&l.MaxTreeDepth, validation.Required.When(l.Apply())),
		validation.Field(&l.MaxEntries, validation.Required.When(l.Apply())),
	)
}

// NoLimits defines no file system FileSystemLimits
func NoLimits() ILimits {
	return &noLimits{}
}

// NewLimits defines file system FileSystemLimits.
func NewLimits(maxFileSize, maxTotalSize, maxFileCount, maxTreeDepth, maxEntries int64) ILimits {
	return &Limits{
		MaxFileSize:  maxFileSize,
		MaxTotalSize: maxTotalSize,
		MaxFileCount: maxFileCount,
		MaxTreeDepth: maxTreeDepth,
		MaxEntries:   maxEntries,
	}
}

// DefaultLimits defines default file system FileSystemLimits
func DefaultLimits() ILimits {
	return NewLimits(1e8, 1e9, 1e6, 12, 1e8) // 100MB, 1GB, 1 million, 12, 100 million
}
