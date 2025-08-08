/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/keyring"
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

// LoadFromSystem is similar to Load but also fetches values from system's [keyring service](https://github.com/zalando/go-keyring).
func LoadFromSystem(envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) error {
	return LoadFromViperAndSystem(viper.New(), envVarPrefix, configurationToSet, defaultConfiguration)
}

// LoadFromViperAndSystem is the same as `LoadFromViper` but also fetches values from system's [keyring service](https://github.com/zalando/go-keyring).
func LoadFromViperAndSystem(viperSession *viper.Viper, envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) error {
	return LoadFromEnvironmentAndSystem(viperSession, envVarPrefix, configurationToSet, defaultConfiguration, "", true)
}

// LoadFromViper is the same as `Load` but instead of creating a new viper session, reuse the one provided.
// Important note:
// Viper's precedence order is maintained:
// 1) values set using explicit calls to `Set`
// 2) flags
// 3) environment (variables or `.env`)
// 4) key/value store
// 5) default values (set via flag default values, or calls to `SetDefault` or via `defaultConfiguration` argument provided)
// Nonetheless, when it comes to default values. It differs slightly from Viper as default values from the default Configuration (i.e. `defaultConfiguration` argument provided) will take precedence over defaults set via `SetDefault` or flags unless they are considered empty values according to `reflection.IsEmpty`.
func LoadFromViper(viperSession *viper.Viper, envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration) error {
	return LoadFromEnvironment(viperSession, envVarPrefix, configurationToSet, defaultConfiguration, "")
}

// LoadFromEnvironment is the same as `LoadFromViper` but also gives the ability to load the configuration from a configuration file as long as the format is supported by [Viper](https://github.com/spf13/viper#what-is-viper)
// Important note:
// Viper's precedence order is maintained:
// 1) values set using explicit calls to `Set`
// 2) flags
// 3) environment (variables or `.env`)
// 4) configuration file
// 5) key/value store
// 6) default values (set via flag default values, or calls to `SetDefault` or via `defaultConfiguration` argument provided)
// Nonetheless, when it comes to default values. It differs slightly from Viper as default values from the default Configuration (i.e. `defaultConfiguration` argument provided) will take precedence over defaults set via `SetDefault` or flags unless they are considered empty values according to `reflection.IsEmpty`.
func LoadFromEnvironment(viperSession *viper.Viper, envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration, configFile string) error {
	return LoadFromEnvironmentAndSystem(viperSession, envVarPrefix, configurationToSet, defaultConfiguration, configFile, false)
}

// LoadFromEnvironmentAndSystem is the same as `LoadFromEnvironment` but also gives the ability to load the configuration from system's [keyring service](https://github.com/zalando/go-keyring).
// Important note:
// Viper's precedence order is mostly maintained:
// 1) values defined in keyring (if not empty and keyring is selected - this is the only difference from Viper)
// 2) values set using explicit calls to `Set`
// 3) flags
// 4) environment (variables or `.env`)
// 5) configuration file
// 6) key/value store
// 7) default values (set via flag default values, or calls to `SetDefault` or via `defaultConfiguration` argument provided)
// Nonetheless, when it comes to default values. It differs slightly from Viper as default values from the default Configuration (i.e. `defaultConfiguration` argument provided) will take precedence over defaults set via `SetDefault` or flags unless they are considered empty values according to `reflection.IsEmpty`.
func LoadFromEnvironmentAndSystem(viperSession *viper.Viper, envVarPrefix string, configurationToSet IServiceConfiguration, defaultConfiguration IServiceConfiguration, configFile string, useKeyring bool) (err error) {
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

	if configFile != "" {
		err = LoadFromConfigurationFile(viperSession, configFile)
		if err != nil {
			return
		}
	}

	// Merge together all the sources and unmarshal into struct
	err = viperSession.Unmarshal(configurationToSet)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "unable to fill configuration structure from the configuration session")
		return
	}
	if useKeyring {
		err = commonerrors.Ignore(keyring.FetchPointer[IServiceConfiguration](context.Background(), envVarPrefix, configurationToSet), commonerrors.ErrUnsupported)
		if err != nil {
			return
		}
	}
	// Run validation
	err = WrapValidationError(field.ToOptionalString(envVarPrefix), configurationToSet.Validate())
	return
}

// LoadFromConfigurationFile loads the configuration from the environment.
// If the format is not supported, an error is raised and the same happens if the file cannot be found.
// Supported formats are the same as what [viper](https://github.com/spf13/viper#what-is-viper) supports
func LoadFromConfigurationFile(viperSession *viper.Viper, configFile string) (err error) {
	if configFile == "" {
		err = commonerrors.UndefinedVariable("configuration file")
		return
	}
	viperSession.SetConfigFile(configFile)
	err = convertViperError(viperSession.MergeInConfig())
	return
}

func convertViperError(vErr error) (err error) {
	vErr = commonerrors.ConvertContextError(vErr)
	switch {
	case vErr == nil:
	case commonerrors.Any(vErr, commonerrors.ErrTimeout, commonerrors.ErrCancelled):
		err = vErr
	case commonerrors.CorrespondTo(vErr, "unsupported"):
		err = commonerrors.WrapError(commonerrors.ErrUnsupported, vErr, "")
	case commonerrors.CorrespondTo(vErr, "not found"):
		err = commonerrors.WrapError(commonerrors.ErrNotFound, vErr, "")
	case commonerrors.CorrespondTo(vErr, "parsing", "marshaling", "decoding"): //nolint: misspell // errors are written in American English
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, vErr, "")
	default:
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, vErr, "")
	}
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

// BindFlagsToEnv binds a set of pflags to an environment variable.
// Envvar is the environment variable string with or without the prefix envVarPrefix
// It is similar to BindFlagToEnv but can be applied to multiple flags.
// Note: all the flags will have to be of the same type. If more than one flags is changed, the system will pick one at random.
func BindFlagsToEnv(viperSession *viper.Viper, envVarPrefix string, envVar string, flags ...*pflag.Flag) (err error) {

	setEnvOptions(viperSession, envVarPrefix)
	shortKey, cleansedEnvVar := generateEnvVarConfigKeys(envVar, envVarPrefix)

	flagset, err := newMultiFlags(shortKey, flags...)
	if err != nil {
		return
	}
	err = viperSession.BindFlagValue(shortKey, flagset)
	if err != nil {
		return
	}

	err = viperSession.BindEnv(shortKey, cleansedEnvVar)
	return
}

func newMultiFlags(name string, flags ...*pflag.Flag) (f viper.FlagValue, err error) {
	if name == "" {
		err = commonerrors.New(commonerrors.ErrUndefined, "flag set must be associated with a name")
		return
	}
	if len(flags) == 0 {
		err = commonerrors.New(commonerrors.ErrUndefined, "flags must be specified")
		return
	}
	var fTypes []string
	for i := range flags {
		if flags[i] != nil {
			fTypes = append(fTypes, flags[i].Value.Type())
		}
	}
	fTypes = collection.UniqueEntries(fTypes)
	if len(fTypes) != 1 {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "flags in a set can only be of the same type: %v", fTypes)
		return
	}

	f = &multiFlags{
		commonName: name,
		flags:      flags,
	}
	return
}

type multiFlags struct {
	commonName string
	flags      []*pflag.Flag
}

func (m *multiFlags) HasChanged() bool {
	for i := range m.flags {
		flag := m.flags[i]
		if flag != nil && flag.Changed {
			return true
		}
	}
	return false
}

func (m *multiFlags) Name() string {
	return m.commonName
}

func (m *multiFlags) ValueString() string {
	var values []string
	var firstValue string
	for i := range m.flags {
		flag := m.flags[i]
		if flag != nil {
			firstValue = flag.Value.String()
			if flag.Changed {
				values = append(values, flag.Value.String())
			}
		}
	}
	values = collection.UniqueEntries(values)
	if len(values) >= 1 {
		return values[0]
	} else {
		return firstValue
	}
}

func (m *multiFlags) ValueType() string {
	for i := range m.flags {
		flag := m.flags[i]
		if flag != nil {
			vType := flag.Value.Type()
			if vType != "" {
				return vType
			}
		}
	}
	return ""
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

func flattenDefaultsMap(m map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range m {
		switch child := value.(type) {
		case map[string]interface{}:
			next := flattenDefaultsMap(child)
			for nextKey, nextValue := range next {
				output[strings.ToUpper(fmt.Sprintf("%s_%s", key, nextKey))] = nextValue
			}
		default:
			output[strings.ToUpper(key)] = value
		}
	}
	return output
}

// DetermineConfigurationEnvironmentVariables returns all the environment variables corresponding to a configuration structure as well as all the default values currently set.
func DetermineConfigurationEnvironmentVariables(appName string, configurationToDecode IServiceConfiguration) (defaults map[string]interface{}, err error) {
	withoutPrefix := make(map[string]interface{})
	if reflection.IsEmpty(configurationToDecode) {
		err = commonerrors.UndefinedVariable("configuration to decode")
		return
	}

	err = mapstructure.Decode(configurationToDecode, &withoutPrefix)
	if err != nil {
		return
	}
	withoutPrefix = flattenDefaultsMap(withoutPrefix)

	defaults = make(map[string]interface{})
	for key, value := range withoutPrefix {
		newKey := fmt.Sprintf("%s_%s", strings.ToUpper(appName), key)
		defaults[newKey] = value
	}
	return
}
