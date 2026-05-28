package frontmatter

import "github.com/ARM-software/golang-utils/utils/collection"

const (
	bufferCap = 1024
)

// ParserOptions is for configuring the front matter parser.
type ParserOptions struct {
	// BufferCapacity specifies the capacity of the buffer used to read the front
	// matter.
	BufferCapacity int
}

// ParserOption is for configuring the front matter parser.
type ParserOption func(*ParserOptions) *ParserOptions

// DefaultParserOptions returns default front matter parser options.
func DefaultParserOptions() *ParserOptions {
	return (&ParserOptions{}).Default()
}

// Default applies default configuration to the front matter parser options.
func (o *ParserOptions) Default() *ParserOptions {
	if o == nil {
		o = &ParserOptions{}
	}
	if o.BufferCapacity <= 0 {
		o.BufferCapacity = bufferCap
	}
	return o
}

// Apply applies an option to the front matter parser options.
func (o *ParserOptions) Apply(option ParserOption) *ParserOptions {
	if option == nil {
		return o.Default()
	}
	return option(o).Default()
}

// NewParserOptions constructs front matter parser options.
func NewParserOptions(options ...ParserOption) *ParserOptions {
	parserOptions := DefaultParserOptions()
	collection.ForEach(options, func(option ParserOption) {
		parserOptions = parserOptions.Apply(option)
	})
	return parserOptions.Default()
}

// WithBufferCapacity is for configuring the front matter parser.
func WithBufferCapacity(capacity int) ParserOption {
	return func(options *ParserOptions) *ParserOptions {
		if options == nil {
			options = DefaultParserOptions()
		}
		options.BufferCapacity = capacity
		return options
	}
}
