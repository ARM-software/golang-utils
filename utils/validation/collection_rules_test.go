package validation

import (
	"iter"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestCollectionRules(t *testing.T) {
	t.Run("array items", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]any{"a", "b"}, ArrayItems(Type("string"))))
		assert.Error(t, validation.Validate([]any{"a", 1}, ArrayItems(Type("string"))))
		errortest.AssertErrorDescription(t, validation.Validate("abc", ArrayItems(Type("string"))), "must be an array or slice")

		var s *string
		errortest.AssertErrorDescription(t, validation.Validate(s, ArrayItems(Type("string"))), "must be an array or slice")

		var f func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(f, ArrayItems(Type("string"))), "must be an array or slice")

		var m map[int]string
		errortest.AssertErrorDescription(t, validation.Validate(m, ArrayItems(Type("string"))), "must be an array or slice")

		seq := iter.Seq[any](func(yield func(any) bool) {
			_ = yield("a")
			_ = yield("b")
		})
		assert.NoError(t, validation.Validate(seq, ArrayItems(Type("string"))))

		var nilSeq iter.Seq[any]
		assert.NoError(t, validation.Validate(nilSeq, ArrayItems(Type("string"))))
	})
}
