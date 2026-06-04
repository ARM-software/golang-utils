//nolint:misspell // The package path intentionally uses "serialization".
package frontmatter

import (
	"context"
	"io"
)

//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/serialization/$GOPACKAGE IFrontMatterParser
//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/serialization/$GOPACKAGE IFrontMatterParser

// IFrontMatterParser extracts front matter content from a reader.
//
// Implementations detect one of the configured front matter formats at the
// start of the stream and return only the extracted front matter content.
type IFrontMatterParser interface {
	// Configure applies a parser option to the front matter parser.
	Configure(...ParserOption) IFrontMatterParser

	// Parse extracts the front matter from r and returns a reader containing only
	// the extracted front matter content.
	Parse(context.Context, io.Reader) (frontmatter io.Reader, err error)

	// Extract extracts the front matter from r and returns the extracted front
	// matter content as raw bytes.
	Extract(context.Context, io.Reader) (frontmatter []byte, err error)
}
