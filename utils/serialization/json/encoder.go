package json

import (
	"context"
	"io"

	"github.com/mailru/easyjson"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/ARM-software/golang-utils/utils/safeio"
)

// Encoder writes successive JSON values to a writer.
// It matches encoding/json.Encoder but wraps the supplied writer so writes obey
// context cancellation, and it uses generated fast JSON encoders when the
// value supports those non-reflective encoding paths.
// It allows encoding many objects to a single writer.
// This should not be used by more than one goroutine at a time.
type Encoder struct {
	writer io.Writer
}

// NewEncoder creates an Encoder that writes JSON values to w.
// It matches encoding/json.NewEncoder.
//
// It keeps the standard Encoder pattern while adding context-aware writes and
// automatic use of generated fast JSON serialisers when a type provides them.
// Examples of supported generators are github.com/mailru/easyjson and
// github.com/pquerna/ffjson.
func NewEncoder(ctx context.Context, w io.Writer) *Encoder {
	return &Encoder{writer: safeio.ContextualWriter(ctx, w)}
}

// Encode writes v as the next JSON value to the encoder's writer.
// It matches encoding/json.Encoder.Encode.
//
// If v implements a generated fast JSON marshaler interface, Encode uses that
// non-reflective path. Otherwise, it falls back to the regular runtime path
// available through the supported fast JSON backend. This preserves a familiar
// API for callers while making fast serialisers an internal optimisation
// instead of a caller concern. Examples of supported generators are
// github.com/mailru/easyjson and github.com/pquerna/ffjson.
func (e *Encoder) Encode(v any) error {
	fastMarshaller, ok := v.(easyjson.Marshaler)
	if ok {
		_, err := easyjson.MarshalToWriter(fastMarshaller, e.writer)
		return err
	}

	return ffjson.NewEncoder(e.writer).Encode(v)
}
