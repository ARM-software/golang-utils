/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
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
	configKeySeparator = "."
	DotEnvFile         = ".env"
)

// Load loads the configuration from the environment (i.e. .env file, environment variables) and puts the entries into the configuration object configurationToSet.
// If not found in the environment, the values will come from the default values defined in defaultConfiguration.
// `envVarPrefix` defines a prefix that ENVIRONMENT variables will use.  E.g. if your prefix is "spf", the env registry will look for env variables that start with "SPF_".
// make sure that the tags on the fields of configurationToSet are properly set using only `[_1-9a-zA-Z]` characters.
func Load(envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) error {
	return LoadFromViper(viper.New(), envVarPrefix, configurationToSet, defaultConfiguration)
}

// LoadFromViper is the same as `Load` but instead of creating a new viper session, reuse the one provided.
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

	// Creating aliases for flags/environment variable keys to real structure keys.
	keys := viperSession.AllKeys()
	for i := range keys {
		key := keys[i]
		longFlagKey, shortFlagKey := generateEnvVarConfigKeys(key, envVarPrefix)

		// The following is a workaround of the aliases implementation in viper which does not really work well with multiple level keys
		value := viperSession.Get(shortFlagKey)
		//Note: Have to use a `IsEmpty` function because the `IsSet` in viper does not consider the default values of a flag.
		if reflection.IsEmpty(value) {
			value = viperSession.Get(longFlagKey)
			if !reflection.IsEmpty(value) {
				viperSession.Set(key, value)
			}
		} else {
			viperSession.Set(key, value)
		}

		viperSession.RegisterAlias(shortFlagKey, key)
		viperSession.RegisterAlias(longFlagKey, key)
	}

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
	longKey, shortKey := generateEnvVarConfigKeys(envVar, envVarPrefix)
	err = viperSession.BindPFlag(longKey, flag)
	if err != nil {
		return
	}

	err = viperSession.BindPFlag(shortKey, flag)
	if err != nil {
		return
	}

	err = viperSession.BindEnv(shortKey)
	return
}

func generateEnvVarConfigKeys(envVar, envVarPrefix string) (longKey, shortKey string) {
	envVarLower := strings.ToLower(envVar)
	envVarPrefixLower := strings.ToLower(envVarPrefix)
	hasPrefix := strings.HasPrefix(envVarLower, envVarPrefixLower)
	var short, extended string
	if hasPrefix {
		extended = envVarLower
		short = strings.TrimPrefix(strings.TrimPrefix(envVarLower, envVarPrefixLower), "_")
	} else {
		extended = fmt.Sprintf("%v_%v", envVarPrefixLower, envVarLower)
		short = strings.ToLower(envVar)
	}

	longKey = generateEnvVarConfigKey(extended)
	shortKey = generateEnvVarConfigKey(short)
	return
}

func generateEnvVarConfigKey(EnvVar string) (key string) {
	key = strings.NewReplacer(EnvVarSeparator, configKeySeparator).Replace(EnvVar)
	return
}

func setEnvOptions(viperSession *viper.Viper, envVarPrefix string) {
	viperSession.SetEnvPrefix(envVarPrefix)
	viperSession.AllowEmptyEnv(false)

	viperSession.AutomaticEnv()
	viperSession.SetEnvKeyReplacer(strings.NewReplacer(configKeySeparator, EnvVarSeparator))
}
