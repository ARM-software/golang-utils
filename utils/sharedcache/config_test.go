package sharedcache

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/require"
)

func TestDefaultSharedCacheConfiguration(t *testing.T) {
	cfg := DefaultSharedCacheConfiguration()
	require.Error(t, cfg.Validate())
	cfg.RemoteStoragePath = faker.URL()
	require.NoError(t, cfg.Validate())
}
