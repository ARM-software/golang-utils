package yaml

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/safeio"
)

// Encoder writes YAML values to a writer.
// It follows the same helper shape as serialization/json.Encoder.
//
// Writes are context-aware. Values are first encoded through the JSON helpers
// so any generated fast JSON serialisers can still be used before converting
// the JSON output to YAML.
// This should not be used by more than one goroutine at a time.
type Encoder struct {
	writer io.Writer
}

// NewEncoder creates an Encoder that writes YAML values to w.
// It follows the same helper shape as serialization/json.NewEncoder.
func NewEncoder(ctx context.Context, w io.Writer) *Encoder {
	return &Encoder{writer: safeio.ContextualWriter(ctx, w)}
}

// Encode writes v as YAML to the encoder's writer.
// It follows the same helper shape as serialization/json.Encoder.Encode.
//
// The value is first encoded through the JSON helpers, then the JSON output is
// converted to YAML before it is written.
func (e *Encoder) Encode(v any) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}

	_, err = e.writer.Write(data)
	return err
}
