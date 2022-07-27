/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	EnvVarSeparator    = "_"
	DotEnvFile         = ".env"
	configKeySeparator = "."
	flagKeyPrefix      = "uniqueprefixforprivateflagbindingkeys123" // Has to be lower case and hopefully unique
)

// Load loads the configuration from the environment (i.e. .env file, environment variables) and puts the entries into the configuration object configurationToSet.
// If not found in the environment, the values will come from the default values defined in defaultConfiguration.
// `envVarPrefix` defines a prefix that ENVIRONMENT variables will use.  E.g. if your prefix is "spf", the env registry will look for env variables that start with "SPF_".
// make sure that the tags on the fields of configurationToSet are properly set using only `[_1-9a-zA-Z]` characters.
func Load(envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) error {
	return LoadFromViper(viper.New(), envVarPrefix, configurationToSet, defaultConfiguration)
}

// LoadFromViper is the same as `Load` but instead of creating a new viper session, reuse the one provided.
// Important note:
// Viper's precedence order is maintained:
//    1) values set using explicit calls to `Set`
//    2) flags
//    3) environment (variables or `.env`)
//    4) configuration file
//    5) key/value store
//    6) default values (set via flag default values, or calls to `SetDefault` or via `defaultConfiguration` argument provided)
// Nonetheless, when it comes to default values. It differs slightly from Viper as default values from the default Configuration (i.e. `defaultConfiguration` argument provided) will take precedence over defaults set via `SetDefault` or flags unless they are considered empty values according to `reflection.IsEmpty`.
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
	_ = godotenv.Load(DotEnvFile)

	// Load Environment variables
	setEnvOptions(viperSession, envVarPrefix)

	linkFlagKeysToStructureKeys(viperSession, envVarPrefix)

	// Merge together all the sources and unmarshal into struct
	err = viperSession.Unmarshal(configurationToSet)
	if err != nil {
		err = fmt.Errorf("unable to decode config into struct, %w", err)
		return
	}
	// Run validation
	err = configurationToSet.Validate()
	return
}

// BindFlagToEnv binds pflags to environment variable.
// Envvar is the environment variable string with or without the prefix envVarPrefix
func BindFlagToEnv(viperSession *viper.Viper, envVarPrefix string, envVar string, flag *pflag.Flag) (err error) {
	setEnvOptions(viperSession, envVarPrefix)
	shortKey, cleansedEnvVar := generateEnvVarConfigKeys(envVar, envVarPrefix)

	err = viperSession.BindPFlag(shortKey, flag)
	if err != nil {
		return
	}
	err = viperSession.BindEnv(shortKey, cleansedEnvVar)
	return
}

func generateEnvVarConfigKeys(envVar, envVarPrefix string) (shortKey string, cleansedEnvVar string) {
	envVarLower := strings.ToLower(envVar)
	envVarPrefixLower := strings.ToLower(envVarPrefix)
	hasPrefix := strings.HasPrefix(envVarLower, envVarPrefixLower)
	var short string
	if hasPrefix {
		short = strings.TrimPrefix(strings.TrimPrefix(envVarLower, envVarPrefixLower), EnvVarSeparator)
	} else {
		short = strings.ToLower(envVar)
	}
	shortKey = generateEnvVarConfigKey(short)
	cleansedEnvVar = cleanseEnvVar(envVarPrefix, short)
	return
}

func generateEnvVarConfigKey(shortEnvVar string) (key string) {
	key = fmt.Sprintf("%v%v%v", flagKeyPrefix, configKeySeparator, strings.NewReplacer(EnvVarSeparator, configKeySeparator).Replace(shortEnvVar))
	return
}

func cleanseEnvVar(envVarPrefix string, shortEnvVar string) (cleansedEnvVar string) {
	cleansedEnvVar = strings.ToUpper(strings.NewReplacer(configKeySeparator, EnvVarSeparator).Replace(fmt.Sprintf("%v%v%v", envVarPrefix, EnvVarSeparator, shortEnvVar)))
	return
}

func isFlagKey(key string) bool {
	return strings.HasPrefix(key, flagKeyPrefix)
}

func setEnvOptions(viperSession *viper.Viper, envVarPrefix string) {
	viperSession.SetEnvPrefix(envVarPrefix)
	viperSession.AllowEmptyEnv(false)

	viperSession.AutomaticEnv()
	viperSession.SetEnvKeyReplacer(strings.NewReplacer(configKeySeparator, EnvVarSeparator))
}

// linkFlagKeysToStructureKeys creates aliases for flags/environment variable keys to real structure keys.
// It was indeed noticed that viper binding/aliasing did not work well with structured/nested configurations.
// Therefore, binding between flags and structure configurations is manually handled.
func linkFlagKeysToStructureKeys(viperSession *viper.Viper, envVarPrefix string) {
	// The following is a workaround of the aliases implementation in viper which does not really work well with multi-level keys
	// Similarly BindEnv does not seem to work well with multi-level configuration structures
	keys := viperSession.AllKeys()
	for i := range keys {
		key := keys[i]
		// This is modifying the value of the structured configuration if flags have been set.
		if !isFlagKey(key) {
			flagKey, _ := generateEnvVarConfigKeys(key, envVarPrefix)
			// if the flag is set, it takes precedence over the structured configuration value.
			if viperSession.IsSet(flagKey) {
				viperSession.Set(key, viperSession.Get(flagKey))
			} else {
				value := viperSession.Get(flagKey)
				if !reflection.IsEmpty(value) {
					viperSession.SetDefault(key, value)
					// If the value of the structured configuration is empty, default to the default value of the flag.
					if reflection.IsEmpty(viperSession.Get(key)) {
						viperSession.Set(key, value)
					}
				}
			}
			viperSession.RegisterAlias(flagKey, key)
		}
	}
}
