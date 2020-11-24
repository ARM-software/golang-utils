package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type IServiceConfiguration interface {
	// Validates configuration entries.
	Validate() error
}

// Loads the configuration from the environment and puts the entries into the configuration object.
// If not found in the environment, the values will come from the default values.
// `envVarPrefix` defines a prefix that ENVIRONMENT variables will use.  E.g. if your prefix is "spf", the env registry will look for env variables that start with "SPF_".
func Load(envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) (err error) {
	v := viper.New()

	// Load Defaults
	var defaults map[string]interface{}
	err = mapstructure.Decode(defaultConfiguration, &defaults)
	if err != nil {
		return
	}
	err = v.MergeConfigMap(defaults)
	if err != nil {
		return
	}

	// Load .env file contents into environment, if it exists
	_ = godotenv.Load(".env")

	// Load Environment variables
	v.SetEnvPrefix(envVarPrefix)
	v.AllowEmptyEnv(false)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Merge together all the sources and unmarshal into struct
	if err := v.Unmarshal(configurationToSet); err != nil {
		return fmt.Errorf("unable to decode config into struct, %w", err)
	}
	// Run validation
	err = configurationToSet.Validate()
	return
}
