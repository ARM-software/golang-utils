package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type IServiceConfiguration interface {
	// Validates configuration entries.
	Validate() error
}

// Loads the configuration from the environment and puts the entries into the configuration object.
// If not found in the environment, the values will come from the default values.
// `envVarPrefix` defines a prefix that ENVIRONMENT variables will use.  E.g. if your prefix is "spf", the env registry will look for env variables that start with "SPF_".
func Load(envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) error {
	return LoadFromViper(viper.New(), envVarPrefix, configurationToSet, defaultConfiguration)

}

// Same as `Load` but instead of creating a new viper session, reuse the one provided.
func LoadFromViper(viperSession *viper.Viper, envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) (err error) {
	// Load Defaults
	var defaults map[string]interface{}
	err = mapstructure.Decode(defaultConfiguration, &defaults)
	if err != nil {
		return
	}
	err = viperSession.MergeConfigMap(defaults)
	if err != nil {
		return
	}

	// Load .env file contents into environment, if it exists
	_ = godotenv.Load(".env")

	// Load Environment variables
	viperSession.SetEnvPrefix(envVarPrefix)
	viperSession.AllowEmptyEnv(false)
	viperSession.AutomaticEnv()
	viperSession.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Merge together all the sources and unmarshal into struct
	if err := viperSession.Unmarshal(configurationToSet); err != nil {
		return fmt.Errorf("unable to decode config into struct, %w", err)
	}
	// Run validation
	err = configurationToSet.Validate()
	return
}

// Binds pflags to environment variable.
func BindFlagToEnv(viperSession *viper.Viper, envVarPrefix string, envVar string, flag *pflag.Flag) (err error) {
	err = viperSession.BindPFlag(envVar, flag)
	if err != nil {
		return
	}
	trimmed := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(envVar), strings.ToLower(envVarPrefix)), "_")
	err = viperSession.BindPFlag(trimmed, flag)
	if err != nil {
		return
	}
	err = viperSession.BindPFlag(strings.ReplaceAll(trimmed, "_", "."), flag)
	if err != nil {
		return
	}
	err = viperSession.BindPFlag(strings.ReplaceAll(envVar, "_", "."), flag)
	return
}
