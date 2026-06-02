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

// NewParser returns a parser for the supplied front matter formats.
//
// The formats are checked in order, with all formats sharing the detected
// opening delimiter considered during the same parsing pass.
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

	return &parser{formats: validatedFormats, options: DefaultParserOptions()}, nil
}

// NewParserWithFormatOptions returns a parser using a Format built from the
// supplied format options.
func NewParserWithFormatOptions(options ...FormatOption) (IFrontMatterParser, error) {
	return NewParser(NewFormat(options...))
}

type parser struct {
	formats []*Format
	options *ParserOptions
}

// Configure applies a parser option to the parser.
func (p *parser) Configure(options ...ParserOption) IFrontMatterParser {
	if p == nil {
		return (&parser{options: DefaultParserOptions()}).Configure(options...)
	}
	if p.options == nil {
		p.options = DefaultParserOptions()
	}
	collection.ForEach(options, func(option ParserOption) {
		p.options = p.options.Apply(option)
	})
	return p
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
		output:  bytes.NewBuffer(make([]byte, 0, p.options.BufferCapacity)),
	}

	content, err := extractor.extract()
	if err != nil {
		return
	}

	frontmatter = safeio.NewByteReader(ctx, content)
	return
}

type candidateFormat struct {
	format *Format
	start  int
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
	candidates, found, err := e.detect()
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, commonerrors.New(commonerrors.ErrNotFound, "front matter not found")
	}

	start, end, found, err := e.capture(candidates)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, commonerrors.New(commonerrors.ErrNotFound, "front matter not found")
	}

	e.start = start
	e.end = end

	return e.output.Bytes()[start:end], nil
}

func (e *extractor) detect() (candidates []*candidateFormat, found bool, err error) {
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

			candidateStart := read
			if !format.IncludeDelimitersWhenUnmarshalling {
				candidateStart = e.read
			}

			candidates = append(candidates, &candidateFormat{format: format, start: candidateStart})
		}

		if len(candidates) == 0 {
			return nil, false, nil
		}

		return candidates, true, nil
	}
}

func (e *extractor) capture(candidates []*candidateFormat) (start int, end int, found bool, err error) {
	for {
		read := e.read

		line, err := e.readLine()
		if commonerrors.Ignore(err, commonerrors.ErrEOF) != nil {
			return 0, 0, false, err
		}

		matchingCandidates := matchingEndCandidates(candidates, line)
		if len(matchingCandidates) == 0 {
			if err != nil {
				return 0, 0, false, nil
			}

			continue
		}

		for _, candidate := range matchingCandidates {
			if candidate.format.RequiresNewLine {
				continue
			}

			candidateEnd := read
			if candidate.format.IncludeDelimitersWhenUnmarshalling {
				candidateEnd = e.read
			}

			return candidate.start, candidateEnd, true, nil
		}

		endAfterClosingLine := read
		if matchingCandidates[0].format.IncludeDelimitersWhenUnmarshalling {
			endAfterClosingLine = e.read
		}

		nextLine, nextErr := e.readLine()
		nextEOF := commonerrors.Any(nextErr, commonerrors.ErrEOF)
		if commonerrors.Ignore(nextErr, commonerrors.ErrEOF) != nil {
			return 0, 0, false, nextErr
		}

		if nextLine != "" {
			if nextEOF {
				return 0, 0, false, nil
			}

			continue
		}

		candidate := matchingCandidates[0]
		candidateEnd := endAfterClosingLine
		if !candidate.format.IncludeDelimitersWhenUnmarshalling {
			candidateEnd = read
		}

		return candidate.start, candidateEnd, true, nil
	}
}

func matchingEndCandidates(candidates []*candidateFormat, line string) (matching []*candidateFormat) {
	for _, candidate := range candidates {
		if candidate.format.End != line {
			continue
		}

		matching = append(matching, candidate)
	}

	return
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
