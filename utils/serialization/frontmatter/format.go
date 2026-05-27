package frontmatter

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// Format describes how a front matter block is detected and extracted.
//
// A format defines the opening and closing delimiters of the front matter and
// how much of the surrounding text should be returned as front matter content.
// This allows callers to model common formats such as YAML, TOML, or JSON front
// matter, as well as custom variants.
// This structure was inspired by what is done in https://github.com/adrg/frontmatter.
type Format struct {
	// Start defines the starting delimiter of the front matter, such as `---`.
	Start string

	// End defines the ending delimiter of the front matter, such as `---`.
	End string

	// UnmarshalDelims specifies whether the returned front matter content should
	// include the start and end delimiters.
	UnmarshalDelims bool

	// RequiresNewLine specifies whether an empty line must follow the ending
	// delimiter for the front matter to be considered complete.
	RequiresNewLine bool
}

// FormatOption configures a front matter Format during construction.
type FormatOption func(*Format) *Format

// DefaultFormat returns a Format initialised with the package defaults.
func DefaultFormat() *Format {
	return (&Format{}).Default()
}

// Default applies the package defaults to a Format.
func (f *Format) Default() *Format {
	if reflection.IsEmpty(f) {
		f = &Format{}
	}
	return f
}

// Apply applies a single option to a Format and then reapplies defaults.
func (f *Format) Apply(option FormatOption) *Format {
	if reflection.IsEmpty(option) {
		return f.Default()
	}
	return option(f).Default()
}

// NewFormat constructs a Format from the supplied options.
func NewFormat(options ...FormatOption) *Format {
	format := DefaultFormat()
	collection.ForEach(options, func(option FormatOption) {
		format = format.Apply(option)
	})
	return format.Default()
}

// Validate checks that the format contains the fields required to detect front
// matter boundaries.
func (f *Format) Validate() error {
	if f == nil {
		return commonerrors.UndefinedVariable("format")
	}
	err := config.ValidateEmbedded(f)
	if err == nil {
		err = validation.ValidateStruct(f,
			validation.Field(&f.Start, validation.Required),
			validation.Field(&f.End, validation.Required),
		)
	}

	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid front matter format")
	}
	return nil
}

// WithStart sets the opening delimiter of a front matter Format.
func WithStart(start string) FormatOption {
	return func(format *Format) *Format {
		if format == nil {
			format = DefaultFormat()
		}
		format.Start = start
		return format
	}
}

// WithEnd sets the closing delimiter of a front matter Format.
func WithEnd(end string) FormatOption {
	return func(format *Format) *Format {
		if format == nil {
			format = DefaultFormat()
		}
		format.End = end
		return format
	}
}

// WithUnmarshalDelimiters specifies to include the start and end delimiters.
func WithUnmarshalDelimiters() FormatOption {
	return func(format *Format) *Format {
		if format == nil {
			format = DefaultFormat()
		}
		format.UnmarshalDelims = true
		return format
	}
}

// WithRequiresNewLine specifies an empty line must follow the closing
// delimiter of the front matter.
func WithRequiresNewLine() FormatOption {
	return func(format *Format) *Format {
		if format == nil {
			format = DefaultFormat()
		}
		format.RequiresNewLine = true
		return format
	}
}
