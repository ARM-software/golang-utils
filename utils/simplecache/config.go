package simplecache

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

type Config struct {
	cachefs                 filesystem.FS `mapstructure:"cache_filesystem"`
	cachePath               string        `mapstructure:"cache_path"`
	garbageCollectionPeriod time.Duration `mapstructure:"gc_period"`
	ttl                     time.Duration `mapstructure:"ttl"`
}

func (cfg *Config) Validate() error {
	// Validate Embedded Structs
	err := configUtils.ValidateEmbedded(cfg)

	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.cachefs, validation.Required),
		validation.Field(&cfg.cachePath, validation.Required),
	)
}
