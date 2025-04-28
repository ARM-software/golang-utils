package filecache

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	configUtils "github.com/ARM-software/golang-utils/utils/config"
)

type FileCacheConfig struct {
	CachePath               string        `mapstructure:"cache_path"`
	GarbageCollectionPeriod time.Duration `mapstructure:"gc_period"`
	TTL                     time.Duration `mapstructure:"ttl"`
}

func (cfg *FileCacheConfig) Validate() error {
	// Validate Embedded Structs
	err := configUtils.ValidateEmbedded(cfg)

	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.CachePath, validation.Required),
		validation.Field(&cfg.GarbageCollectionPeriod, validation.Required),
		validation.Field(&cfg.TTL, validation.Required),
	)
}

func DefaultFileCacheConfig() *FileCacheConfig {
	return &FileCacheConfig{
		GarbageCollectionPeriod: 10 * time.Minute,
		TTL:                     2 * time.Hour,
	}
}
