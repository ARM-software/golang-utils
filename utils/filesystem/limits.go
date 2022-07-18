package filesystem

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

func (n *noLimits) GetMaxTotalSize() uint64 {
	return 0
}

func (n *noLimits) Validate() error {
	return nil
}

// Limits defines file system limits
type Limits struct {
	MaxFileSize  int64
	MaxTotalSize uint64
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
	)
}

// NoLimits defines no file system FileSystemLimits
func NoLimits() ILimits {
	return &noLimits{}
}

// NewLimits defines file system FileSystemLimits.
func NewLimits(maxFileSize int64, maxTotalSize uint64) ILimits {
	return &Limits{
		MaxFileSize:  maxFileSize,
		MaxTotalSize: maxTotalSize,
	}
}

// DefaultLimits defines default file system FileSystemLimits
func DefaultLimits() ILimits {
	return NewLimits(1<<30, 10<<30)
}
