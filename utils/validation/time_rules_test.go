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
		assert.NoError(t, validation.Validate("5s", DurationConst(5*time.Second)))
		assert.Error(t, validation.Validate("6s", DurationConst(5*time.Second)))
	})

	t.Run("timestamp", func(t *testing.T) {
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", IsTimestamp))
		assert.Error(t, validation.Validate("2024-01-01", IsTimestamp))
		minTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		maxTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMinimum(minTime)))
		assert.Error(t, validation.Validate("2023-12-31T00:00:00Z", TimestampMinimum(minTime)))
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMaximum(maxTime)))
		assert.Error(t, validation.Validate("2025-01-01T00:00:00Z", TimestampMaximum(maxTime)))
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", TimestampConst(minTime)))
		assert.Error(t, validation.Validate("2024-01-02T00:00:00Z", TimestampConst(minTime)))
	})
}
