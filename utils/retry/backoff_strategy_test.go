package retry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryPolicyConfigurationByBackoffStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy BackoffStrategy
		assertFn func(*testing.T, *RetryPolicyConfiguration)
	}{
		{
			name:     "no backoff",
			strategy: NoBackoff,
			assertFn: func(t *testing.T, cfg *RetryPolicyConfiguration) {
				t.Helper()
				assert.True(t, cfg.Enabled)
				assert.True(t, cfg.RetryAfterDisabled)
				assert.False(t, cfg.BackOffEnabled)
				assert.Zero(t, cfg.RetryWaitMin)
				assert.Zero(t, cfg.RetryWaitMax)
				assert.Zero(t, cfg.RetryMaxJitter)
			},
		},
		{
			name:     "fixed backoff",
			strategy: FixedBackoff,
			assertFn: func(t *testing.T, cfg *RetryPolicyConfiguration) {
				t.Helper()
				assert.True(t, cfg.Enabled)
				assert.True(t, cfg.RetryAfterDisabled)
				assert.True(t, cfg.BackOffEnabled)
				assert.True(t, cfg.LinearBackOffEnabled)
				assert.Equal(t, cfg.RetryWaitMin, cfg.RetryWaitMax)
				assert.Zero(t, cfg.RetryMaxJitter)
			},
		},
		{
			name:     "no backoff but retry after",
			strategy: NoBackoffButRetryAfter,
			assertFn: func(t *testing.T, cfg *RetryPolicyConfiguration) {
				t.Helper()
				assert.True(t, cfg.Enabled)
				assert.False(t, cfg.RetryAfterDisabled)
				assert.False(t, cfg.BackOffEnabled)
				assert.Zero(t, cfg.RetryWaitMin)
				assert.Zero(t, cfg.RetryWaitMax)
				assert.Zero(t, cfg.RetryMaxJitter)
			},
		},
		{
			name:     "fixed backoff with retry after",
			strategy: FixedBackoffOrRetryAfter,
			assertFn: func(t *testing.T, cfg *RetryPolicyConfiguration) {
				t.Helper()
				assert.True(t, cfg.Enabled)
				assert.False(t, cfg.RetryAfterDisabled)
				assert.True(t, cfg.BackOffEnabled)
				assert.True(t, cfg.LinearBackOffEnabled)
				assert.Zero(t, cfg.RetryMaxJitter)
			},
		},
		{
			name:     "linear backoff",
			strategy: LinearBackoff,
			assertFn: func(t *testing.T, cfg *RetryPolicyConfiguration) {
				t.Helper()
				assert.True(t, cfg.Enabled)
				assert.True(t, cfg.BackOffEnabled)
				assert.True(t, cfg.LinearBackOffEnabled)
				assert.NotZero(t, cfg.RetryMaxJitter)
			},
		},
		{
			name:     "exponential backoff",
			strategy: ExponentialBackoff,
			assertFn: func(t *testing.T, cfg *RetryPolicyConfiguration) {
				t.Helper()
				assert.True(t, cfg.Enabled)
				assert.True(t, cfg.BackOffEnabled)
				assert.False(t, cfg.LinearBackOffEnabled)
				assert.NotZero(t, cfg.RetryMaxJitter)
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			cfg := DefaultRetryPolicyConfiguration(test.strategy)
			require.NotNil(t, cfg)
			test.assertFn(t, cfg)
		})
	}

	assert.Equal(t, DefaultNoRetryPolicyConfiguration(), DefaultRetryPolicyConfiguration(BackoffStrategy(99)))
}
