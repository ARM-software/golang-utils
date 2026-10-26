package json

import (
	"context"
	"io"

	"github.com/mailru/easyjson"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/ARM-software/golang-utils/utils/safeio"
)

// Decoder decodes successive JSON values from a reader.
// It matches encoding/json.Decoder but wraps the supplied reader so reads obey
// context cancellation, and it uses generated fast JSON decoders when the
// target type supports those non-reflective decoding paths.
// This should not be used by more than one goroutine at a time.
type Decoder struct {
	reader io.Reader
}

// NewDecoder creates a Decoder that reads JSON values from r.
// It matches encoding/json.NewDecoder.
//
// It keeps the familiar Decoder workflow while adding context-aware reads and
// automatic use of generated fast JSON deserialisers. Examples of supported
// generators are github.com/mailru/easyjson and github.com/pquerna/ffjson.
func NewDecoder(ctx context.Context, r io.Reader) *Decoder {
	return &Decoder{reader: safeio.NewContextualReader(ctx, r)}
}

// Decode reads the next JSON value from the decoder into v.
// It matches encoding/json.Decoder.Decode.
//
// If v implements a generated fast JSON unmarshaler interface, Decode uses
// that non-reflective path. Otherwise, it falls back to the regular runtime
// path available through the supported fast JSON backend. The goal is to
// preserve a standard-library style API while making the faster implementation
// an internal detail. Examples of supported generators are
// github.com/mailru/easyjson and github.com/pquerna/ffjson.
func (d *Decoder) Decode(v any) error {
	fastUnmarshaller, ok := v.(easyjson.Unmarshaler)
	if ok {
		return easyjson.UnmarshalFromReader(d.reader, fastUnmarshaller)
	}

	return ffjson.NewDecoder().DecodeReader(d.reader, v)
}
