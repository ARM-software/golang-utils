package filecache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileCacheConfig_Validate(t *testing.T) {
	t.Run("empty config requires cache path", func(t *testing.T) {
		cfg := &FileCacheConfig{}
		require.Error(t, cfg.Validate())
	})

	t.Run("zero durations are accepted when cache path is set", func(t *testing.T) {
		cfg := &FileCacheConfig{CachePath: t.TempDir()}
		require.NoError(t, cfg.Validate())
	})

	t.Run("default config is valid once cache path is set", func(t *testing.T) {
		cfg := DefaultFileCacheConfig()
		cfg.CachePath = t.TempDir()
		require.NoError(t, cfg.Validate())
	})
}
