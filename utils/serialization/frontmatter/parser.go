package frontmatter

import (
	"bufio"
	"bytes"
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

const (
	bufferCap = 1024
)

// NewParser returns a parser for the supplied front matter formats.
//
// The formats are checked in order, and the first format whose opening
// delimiter matches the start of the reader is used.
func NewParser(formats ...*Format) (p IFrontMatterParser, err error) {
	if reflection.IsEmpty(formats) {
		err = commonerrors.UndefinedVariable("formats")
		return
	}

	validatedFormats, err := collection.MapWithError[*Format, *Format](formats, func(f *Format) (v *Format, subErr error) {
		if reflection.IsEmpty(f) {
			subErr = commonerrors.UndefinedVariable("format")
			return
		}
		subErr = f.Validate()
		v = f.Default()
		return
	})
	if err != nil {
		return
	}

	return &parser{formats: validatedFormats}, nil
}

// NewParserWithOptions returns a parser using a Format built from the supplied
// options.
func NewParserWithOptions(options ...FormatOption) (IFrontMatterParser, error) {
	return NewParser(NewFormat(options...))
}

type parser struct {
	formats []*Format
}

// Extract extracts the front matter at the start of reader and returns only
// the extracted front matter bytes.
func (p *parser) Extract(ctx context.Context, reader io.Reader) (frontmatter []byte, err error) {
	r, err := p.Parse(ctx, reader)
	if err != nil {
		return
	}
	frontmatter, err = safeio.ReadAll(ctx, r)
	return
}

// Parse extracts the front matter at the start of r and returns a reader that
// contains only the front matter content.
func (p *parser) Parse(ctx context.Context, r io.Reader) (frontmatter io.Reader, err error) {
	if r == nil {
		err = commonerrors.UndefinedVariable("reader")
		return
	}

	extractor := &extractor{
		formats: p.formats,
		reader:  bufio.NewReader(safeio.NewContextualReader(ctx, r)),
		output:  bytes.NewBuffer(make([]byte, 0, bufferCap)),
	}

	content, err := extractor.extract()
	if err != nil {
		return
	}
	frontmatter = safeio.NewByteReader(ctx, content)
	return
}

type extractor struct {
	formats []*Format
	reader  *bufio.Reader
	output  *bytes.Buffer
	read    int
	start   int
	end     int
}

func (e *extractor) extract() ([]byte, error) {
	format, found, err := e.detect()
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, commonerrors.New(commonerrors.ErrNotFound, "front matter not found")
	}

	found, err = e.capture(format)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, commonerrors.New(commonerrors.ErrNotFound, "front matter not found")
	}

	return e.output.Bytes()[e.start:e.end], nil
}

func (e *extractor) detect() (*Format, bool, error) {
	for {
		read := e.read

		line, err := e.readLine()

		if commonerrors.Ignore(err, commonerrors.ErrEOF) != nil {
			return nil, false, err
		}
		if err != nil && line == "" {
			return nil, false, nil
		}
		if line == "" {
			if err != nil {
				return nil, false, nil
			}
			continue
		}

		for _, format := range e.formats {
			if format.Start != line {
				continue
			}
			if !format.UnmarshalDelims {
				read = e.read
			}
			e.start = read
			return format, true, nil
		}

		return nil, false, nil
	}
}

func (e *extractor) capture(format *Format) (bool, error) {
	for {
		read := e.read

		line, err := e.readLine()
		if commonerrors.Ignore(err, commonerrors.ErrEOF) != nil {
			return false, err
		}

		if line != format.End {
			if err != nil {
				return false, nil
			}
			continue
		}

		if format.UnmarshalDelims {
			read = e.read
		}

		if format.RequiresNewLine {
			nextLine, nextErr := e.readLine()
			nextEOF := commonerrors.Any(nextErr, commonerrors.ErrEOF)
			if commonerrors.Ignore(nextErr, commonerrors.ErrEOF) != nil {
				return false, nextErr
			}
			if nextLine != "" {
				if nextEOF {
					return false, nil
				}
				continue
			}
		}

		e.end = read
		return true, nil
	}
}

func (e *extractor) readLine() (line string, err error) {
	lineB, err := e.reader.ReadBytes('\n')
	err = safeio.ConvertIOError(err)

	lineSize := len(lineB)
	if lineSize == 0 {
		return
	}

	e.read += lineSize
	_, err = e.output.Write(lineB)
	line = string(bytes.TrimSpace(lineB))
	return
}
