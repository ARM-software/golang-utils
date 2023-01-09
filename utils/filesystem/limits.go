package filesystem

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/units/multiplication"
	"github.com/ARM-software/golang-utils/utils/units/size"
)

type noLimits struct {
}

func (n *noLimits) GetMaxDepth() int64 {
	return 0
}

func (n *noLimits) ApplyRecursively() bool {
	return false
}

func (n *noLimits) Apply() bool {
	return false
}

func (n *noLimits) GetMaxFileSize() int64 {
	return 0
}

func (n *noLimits) GetMaxTotalSize() uint64 {
	return 0
}

func (n *noLimits) GetMaxFileCount() int64 {
	return 0
}

func (n *noLimits) Validate() error {
	return nil
}

// Limits defines file system limits
type Limits struct {
	MaxFileSize  int64  `mapstructure:"max_file_size"`
	MaxTotalSize uint64 `mapstructure:"max_total_size"`
	MaxFileCount int64  `mapstructure:"max_file_count"`
	MaxDepth     int64  `mapstructure:"max_depth"`
	Recursive    bool   `mapstructure:"recursive"`
}

func (l *Limits) ApplyRecursively() bool {
	return l.Recursive
}

func (l *Limits) Apply() bool {
	return true
}

func (l *Limits) GetMaxFileSize() int64 {
	return l.MaxFileSize
}

func (l *Limits) GetMaxTotalSize() uint64 {
	return l.MaxTotalSize
}

func (l *Limits) GetMaxFileCount() int64 {
	return l.MaxFileCount
}

func (l *Limits) GetMaxDepth() int64 {
	return l.MaxDepth
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
		validation.Field(&l.MaxDepth, validation.Required.When(l.Apply())),
	)
}

// NoLimits defines no file system FileSystemLimits
func NoLimits() ILimits {
	return &noLimits{}
}

// NewLimits defines file system FileSystemLimits.
func NewLimits(maxFileSize int64, maxTotalSize uint64, maxFileCount int64, maxDepth int64, recursive bool) ILimits {
	return &Limits{
		MaxFileSize:  maxFileSize,
		MaxTotalSize: maxTotalSize,
		MaxFileCount: maxFileCount,
		MaxDepth:     maxDepth,
		Recursive:    recursive,
	}
}

// DefaultLimits defines default file system limits
func DefaultLimits() ILimits {
	return NewLimits(int64(size.GiB), uint64(10*size.GiB), multiplication.Mega, -1, true)
}

// DefaultZipLimits defines default file system limits for handling zips
func DefaultZipLimits() ILimits {
	return RecursiveZipLimits(-1)
}

// RecursiveZipLimits defines file system limits for handling zips recursively
func RecursiveZipLimits(maxDepth int64) ILimits {
	return NewLimits(int64(size.GiB), uint64(10*size.GiB), multiplication.Mega, maxDepth, true)
}

// DefaultNonRecursiveZipLimits defines default file system limits for handling zips
func DefaultNonRecursiveZipLimits() ILimits {
	return NewLimits(int64(size.GiB), uint64(10*size.GiB), multiplication.Mega, 10, false)
}
