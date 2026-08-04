package validation

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
)

func TestTimeRules(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		pointerDuration := field.ToOptionalDuration(10 * time.Second)
		var nilPointerDuration *time.Duration

		assert.NoError(t, validation.Validate("5s", IsDuration))
		assert.NoError(t, validation.Validate(pointerDuration, IsDuration))
		errortest.AssertError(t, validation.Validate("not-a-duration", IsDuration), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(nilPointerDuration, IsDuration), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate("10s", DurationMinimum(5*time.Second)))
		assert.NoError(t, validation.Validate(pointerDuration, DurationMinimum(5*time.Second)))
		errortest.AssertError(t, validation.Validate("1s", DurationMinimum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate([]byte("6s"), DurationExclusiveMinimum(5*time.Second)))
		assert.NoError(t, validation.Validate(field.ToOptionalDuration(6*time.Second), DurationExclusiveMinimum(5*time.Second)))
		errortest.AssertError(t, validation.Validate([]byte("5s"), DurationExclusiveMinimum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate(5*time.Second, DurationMaximum(5*time.Second)))
		assert.NoError(t, validation.Validate(field.ToOptionalDuration(5*time.Second), DurationMaximum(5*time.Second)))
		errortest.AssertError(t, validation.Validate(6*time.Second, DurationMaximum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate("4s", DurationExclusiveMaximum(5*time.Second)))
		assert.NoError(t, validation.Validate(field.ToOptionalDuration(4*time.Second), DurationExclusiveMaximum(5*time.Second)))
		errortest.AssertError(t, validation.Validate("5s", DurationExclusiveMaximum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate("5s", DurationConst(5*time.Second)))
		assert.NoError(t, validation.Validate(field.ToOptionalDuration(5*time.Second), DurationConst(5*time.Second)))
		errortest.AssertError(t, validation.Validate("6s", DurationConst(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)

		errortest.AssertError(t, validation.Validate("not-a-duration", DurationMinimum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate("not-a-duration", DurationExclusiveMinimum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate("not-a-duration", DurationMaximum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate("not-a-duration", DurationExclusiveMaximum(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate("not-a-duration", DurationConst(5*time.Second)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
	})

	t.Run("timestamp", func(t *testing.T) {
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", IsRFC3339Timestamp))
		assert.NoError(t, validation.Validate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), IsRFC3339Timestamp))
		assert.NoError(t, validation.Validate(field.ToOptionalTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), IsRFC3339Timestamp))
		assert.NoError(t, validation.Validate(field.ToOptionalString("2024-01-01T00:00:00Z"), IsRFC3339Timestamp))
		errortest.AssertError(t, validation.Validate("2024-01-01", IsRFC3339Timestamp), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(field.ToOptionalString("2024-01-01"), IsRFC3339Timestamp), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(123, IsRFC3339Timestamp), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(field.ToOptionalInt(123), IsRFC3339Timestamp), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate([]byte("not-a-timestamp"), IsRFC3339Timestamp), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		minTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		maxTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		pointerTime := field.ToOptionalTime(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
		pointerStringTime := field.ToOptionalString("2024-06-01T00:00:00Z")
		var nilPointerTime *time.Time
		var nilPointerString *string
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMinimum(minTime)))
		assert.NoError(t, validation.Validate(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), TimestampMinimum(minTime)))
		errortest.AssertError(t, validation.Validate("2023-12-31T00:00:00Z", TimestampMinimum(minTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate([]byte("2024-01-02T00:00:00Z"), TimestampExclusiveMinimum(minTime)))
		errortest.AssertError(t, validation.Validate([]byte("2024-01-01T00:00:00Z"), TimestampExclusiveMinimum(minTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", TimestampMaximum(maxTime)))
		errortest.AssertError(t, validation.Validate("2025-01-01T00:00:00Z", TimestampMaximum(maxTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate(time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC), TimestampExclusiveMaximum(maxTime)))
		errortest.AssertError(t, validation.Validate(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), TimestampExclusiveMaximum(maxTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate("2024-01-01T00:00:00Z", TimestampConst(minTime)))
		errortest.AssertError(t, validation.Validate("2024-01-02T00:00:00Z", TimestampConst(minTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		assert.NoError(t, validation.Validate(pointerTime, TimestampMinimum(minTime)))
		assert.NoError(t, validation.Validate(pointerStringTime, TimestampMinimum(minTime)))
		assert.NoError(t, validation.Validate(nilPointerTime, NilTimestampOrNotEmpty))
		assert.NoError(t, validation.Validate(nilPointerString, NilTimestampOrNotEmpty))
		assert.NoError(t, validation.Validate("2024-06-01T00:00:00Z", NilTimestampOrNotEmpty))
		assert.NoError(t, validation.Validate(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), NilTimestampOrNotEmpty))
		assert.NoError(t, validation.Validate(pointerTime, NilTimestampOrNotEmpty))
		assert.NoError(t, validation.Validate(pointerStringTime, NilTimestampOrNotEmpty))
		errortest.AssertError(t, validation.Validate("", NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(field.ToOptionalString(""), NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate("not-a-timestamp", NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(field.ToOptionalString("not-a-timestamp"), NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(123, NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(time.Time{}, NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(field.ToOptionalTime(time.Time{}), NilTimestampOrNotEmpty), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate("not-a-timestamp", TimestampMinimum(minTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(field.ToOptionalString("not-a-timestamp"), TimestampMinimum(minTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, validation.Validate(123, TimestampMinimum(minTime)), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
	})
}
