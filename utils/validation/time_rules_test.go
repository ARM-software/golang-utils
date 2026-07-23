package validation

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestTimeRules(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		assert.NoError(t, validation.Validate("5s", IsDuration))
		assert.Error(t, validation.Validate("not-a-duration", IsDuration))
		assert.NoError(t, validation.Validate("10s", DurationMinimum(5*time.Second)))
		assert.Error(t, validation.Validate("1s", DurationMinimum(5*time.Second)))
		assert.NoError(t, validation.Validate([]byte("6s"), DurationExclusiveMinimum(5*time.Second)))
		assert.Error(t, validation.Validate([]byte("5s"), DurationExclusiveMinimum(5*time.Second)))
		assert.NoError(t, validation.Validate(5*time.Second, DurationMaximum(5*time.Second)))
		assert.Error(t, validation.Validate(6*time.Second, DurationMaximum(5*time.Second)))
		assert.NoError(t, validation.Validate("4s", DurationExclusiveMaximum(5*time.Second)))
		assert.Error(t, validation.Validate("5s", DurationExclusiveMaximum(5*time.Second)))
		assert.NoError(t, validation.Validate("5s", DurationConst(5*time.Second)))
		assert.Error(t, validation.Validate("6s", DurationConst(5*time.Second)))

		// Parsing failures currently return nil because each rule exits on !ok.
		assert.Error(t, validation.Validate("not-a-duration", DurationMinimum(5*time.Second)))
		assert.Error(t, validation.Validate("not-a-duration", DurationExclusiveMinimum(5*time.Second)))
		assert.Error(t, validation.Validate("not-a-duration", DurationMaximum(5*time.Second)))
		assert.Error(t, validation.Validate("not-a-duration", DurationExclusiveMaximum(5*time.Second)))
		assert.Error(t, validation.Validate("not-a-duration", DurationConst(5*time.Second)))
	})

	t.Run("timestamp", func(t *testing.T) {
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", IsRFC3339Timestamp))
		assert.Error(t, validation.Validate("2024-01-01", IsRFC3339Timestamp))
		minTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		maxTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMinimum(minTime)))
		assert.Error(t, validation.Validate("2023-12-31T00:00:00Z", TimestampMinimum(minTime)))
		assert.NoError(t, validation.Validate([]byte("2024-01-02T00:00:00Z"), TimestampExclusiveMinimum(minTime)))
		assert.Error(t, validation.Validate([]byte("2024-01-01T00:00:00Z"), TimestampExclusiveMinimum(minTime)))
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMaximum(maxTime)))
		assert.Error(t, validation.Validate("2025-01-01T00:00:00Z", TimestampMaximum(maxTime)))
		assert.NoError(t, validation.Validate(time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC), TimestampExclusiveMaximum(maxTime)))
		assert.Error(t, validation.Validate(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), TimestampExclusiveMaximum(maxTime)))
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", TimestampConst(minTime)))
		assert.Error(t, validation.Validate("2024-01-02T00:00:00Z", TimestampConst(minTime)))
	})
}
