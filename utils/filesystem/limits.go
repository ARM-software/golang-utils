package filesystem

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

type limits struct {
	MaxFileSize  int64
	MaxTotalSize uint64
}

func (l *limits) Apply() bool {
	return true
}

func (l *limits) GetMaxFileSize() int64 {
	return l.MaxFileSize
}

func (l *limits) GetMaxTotalSize() uint64 {
	return l.MaxTotalSize
}

// NoLimits defines no file system limits
func NoLimits() ILimits {
	return &noLimits{}
}

// NewLimits defines file system limits.
func NewLimits(maxFileSize int64, maxTotalSize uint64) ILimits {
	return &limits{
		MaxFileSize:  maxFileSize,
		MaxTotalSize: maxTotalSize,
	}
}

// DefaultLimits defines default file system limits
func DefaultLimits() ILimits {
	return NewLimits(1<<30, 10<<30)
}
