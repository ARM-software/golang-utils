package sharedcache

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	configUtils "github.com/ARM-software/golang-utils/utils/config"
)

type Configuration struct {
	RemoteStoragePath       string        `mapstructure:"remote_storage_path"` // Path where the cache will be stored.
	Timeout                 time.Duration `mapstructure:"timeout"`             // Cache timeout if need be
	FilesystemItemsToIgnore string        `mapstructure:"ignore_fs_items"`     // List of files/folders to ignore (pattern list separated by commas)
}

func (cfg *Configuration) Validate() error {
	// Validate Embedded Structs
	err := configUtils.ValidateEmbedded(cfg)

	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.RemoteStoragePath, validation.Required),
	)
}

func DefaultSharedCacheConfiguration() *Configuration {
	return &Configuration{}
}
