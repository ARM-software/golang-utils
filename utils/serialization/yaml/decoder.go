package yaml

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/safeio"
)

// Decoder reads YAML from a reader and decodes it into Go values.
// It follows the same helper shape as serialization/json.Decoder.
//
// Reads are context-aware, and the input is converted to JSON before the JSON
// helpers are used, so the repository keeps a single fast decoding path.
// This should not be used by more than one goroutine at a time.
type Decoder struct {
	ctx    context.Context
	reader io.Reader
}

// NewDecoder creates a Decoder that reads YAML values from r.
// It follows the same helper shape as serialization/json.NewDecoder.
func NewDecoder(ctx context.Context, r io.Reader) *Decoder {
	return &Decoder{ctx: ctx, reader: safeio.NewContextualReader(ctx, r)}
}

// Decode reads YAML from the decoder and stores the result in v.
// It follows the same helper shape as serialization/json.Decoder.Decode.
//
// The data is converted to JSON first, then passed to the JSON helpers, so the
// same fast and implementation-agnostic decoding logic is reused here.
func (d *Decoder) Decode(v any) error {
	data, err := safeio.ReadAll(d.ctx, d.reader)
	if err != nil {
		return err
	}

	return Unmarshal(data, v)
}
