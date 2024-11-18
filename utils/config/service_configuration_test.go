/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

var (
	random           = rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
	expectedString   = fmt.Sprintf("a test string %v", faker.Word())
	expectedInt      = random.Int() //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	expectedDuration = time.Since(time.Date(1999, 2, 3, 4, 30, 45, 46, time.UTC))
	expectedHost     = fmt.Sprintf("a test host %v", faker.Word())
	expectedPassword = fmt.Sprintf("a test passwd %v", faker.Password())
	expectedDB       = fmt.Sprintf("a db %v", faker.Word())
)

type DummyConfiguration struct {
	Host              string        `mapstructure:"dummy_host"`
	Port              int           `mapstructure:"port"`
	DB                string        `mapstructure:"db"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Flag              bool          `mapstructure:"flag"`
	HealthCheckPeriod time.Duration `mapstructure:"healthcheck_period"`
}

func (cfg *DummyConfiguration) Validate() error {
	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required, validation.Min(0)),
		validation.Field(&cfg.DB, validation.Required),
		validation.Field(&cfg.User, validation.Required),
		validation.Field(&cfg.Password, validation.Required),
	)
}

func DefaultDummyConfiguration() *DummyConfiguration {
	return &DummyConfiguration{
		Port:              5432,
		Flag:              true,
		HealthCheckPeriod: time.Second,
	}
}

type ConfigurationTest struct {
	TestString  string             `mapstructure:"dummy_string"`
	TestInt     int                `mapstructure:"dummy_int"`
	TestTime    time.Duration      `mapstructure:"dummy_time"`
	TestConfig  DummyConfiguration `mapstructure:"dummyconfig"`
	TestConfig2 DummyConfiguration `mapstructure:"dummy_config"`
}

type DeepConfig struct {
	TestString     string            `mapstructure:"dummy_string"`
	TestConfigDeep ConfigurationTest `mapstructure:"deep_config"`
}

func DefaultDeepConfiguration() *DeepConfig {
	return &DeepConfig{
		TestString:     expectedString,
		TestConfigDeep: *DefaultConfiguration(),
	}
}

func (cfg *DeepConfig) Validate() error {
	return nil
}

func (cfg *ConfigurationTest) Validate() error {
	validation.ErrorTag = "mapstructure"

	// Validate Embedded Structs
	err := ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.TestString, validation.Required),
		validation.Field(&cfg.TestInt, validation.Required),
		validation.Field(&cfg.TestTime, validation.Required),
		validation.Field(&cfg.TestConfig, validation.Required),
	)
}

func DefaultConfiguration() *ConfigurationTest {
	return &ConfigurationTest{
		TestString:  expectedString,
		TestInt:     0,
		TestTime:    time.Hour,
		TestConfig:  *DefaultDummyConfiguration(),
		TestConfig2: *DefaultDummyConfiguration(),
	}
}

func TestServiceConfigurationLoad(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	err := Load("test", configTest, defaults)
	// Some required values are missing.
	require.Error(t, err)
	require.NotNil(t, configTest.Validate())
	// Setting required entries in the environment.
	err = os.Setenv("TEST_DUMMYCONFIG_DUMMY_HOST", expectedHost)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_DUMMY_HOST", "a test host")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_PASSWORD", "a test password")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_PASSWORD", expectedPassword)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_USER", "a test user")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_USER", "a test user")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_DB", "a test db")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_DB", expectedDB)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_FLAG", "false")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_TIME", expectedDuration.String())
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_INT", fmt.Sprintf("%v", expectedInt))
	require.NoError(t, err)
	err = Load("test", configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, expectedString, configTest.TestString)
	assert.Equal(t, expectedInt, configTest.TestInt)
	assert.Equal(t, expectedDuration, configTest.TestTime)
	assert.Equal(t, defaults.TestConfig.Port, configTest.TestConfig.Port)
	assert.Equal(t, expectedHost, configTest.TestConfig.Host)
	assert.Equal(t, expectedPassword, configTest.TestConfig2.Password)
	assert.Equal(t, expectedDB, configTest.TestConfig2.DB)
	assert.NotEqual(t, expectedHost, configTest.TestConfig2.Host)
	assert.NotEqual(t, expectedPassword, configTest.TestConfig.Password)
	assert.NotEqual(t, expectedDB, configTest.TestConfig.DB)
	assert.True(t, configTest.TestConfig.Flag)
	assert.False(t, configTest.TestConfig2.Flag)
}

func TestServiceConfigurationLoad_Errors(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	err := Load("test", configTest, DefaultConfiguration())
	// Some required values are missing.
	require.Error(t, err)
	require.NotNil(t, configTest.Validate())

	err = Load("test", nil, DefaultDummyConfiguration())
	// Incorrect  structure provided.
	require.Error(t, err)
}

func TestSimpleFlagBinding(t *testing.T) {
	os.Clearenv()
	configTest := &DummyConfiguration{}
	defaults := DefaultDummyConfiguration()
	session := viper.New()
	var err error
	flagSet := pflag.FlagSet{}
	prefix := "test"
	flagSet.String("host", "a host", "dummy host")
	flagSet.String("password", "a password", "dummy password")
	flagSet.String("user", "a user", "dummy user")
	flagSet.String("db", "a db", "dummy db")
	err = BindFlagToEnv(session, prefix, "TEST_DUMMY_HOST", flagSet.Lookup("host"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "PASSWORD", flagSet.Lookup("password"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "TEST_DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "USER", flagSet.Lookup("user"))
	require.NoError(t, err)
	err = flagSet.Set("host", expectedHost)
	require.NoError(t, err)
	err = flagSet.Set("password", expectedPassword)
	require.NoError(t, err)
	err = flagSet.Set("db", expectedDB)
	require.NoError(t, err)
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, defaults.Port, configTest.Port)
	assert.Equal(t, expectedHost, configTest.Host)
	assert.Equal(t, expectedPassword, configTest.Password)
	assert.Equal(t, expectedDB, configTest.DB)
	assert.True(t, configTest.Flag)
}

func TestFlagBinding(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	session := viper.New()
	var err error
	flagSet := pflag.FlagSet{}
	prefix := "test"
	flagSet.String("host", "a host", "dummy host")
	flagSet.String("password", "a password", "dummy password")
	flagSet.String("user", "a user", "dummy user")
	flagSet.String("db", "a db", "dummy db")
	flagSet.String("db2", "a db", "dummy db")
	flagSet.Int("int", 0, "dummy int")
	flagSet.Duration("time", time.Second, "dummy time")
	flagSet.Bool("flag", false, "dummy flag")
	err = BindFlagToEnv(session, prefix, "TEST_DUMMYCONFIG_DUMMY_HOST", flagSet.Lookup("host"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "TEST_DUMMY_CONFIG_DUMMY_HOST", flagSet.Lookup("host"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMYCONFIG_PASSWORD", flagSet.Lookup("password"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_PASSWORD", flagSet.Lookup("password"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMYCONFIG_USER", flagSet.Lookup("user"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_USER", flagSet.Lookup("user"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "TEST_DUMMYCONFIG_DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_DB", flagSet.Lookup("db2"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_FLAG", flagSet.Lookup("flag"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_INT", flagSet.Lookup("int"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_Time", flagSet.Lookup("time"))
	require.NoError(t, err)
	err = flagSet.Set("host", expectedHost)
	require.NoError(t, err)
	err = flagSet.Set("password", expectedPassword)
	require.NoError(t, err)
	err = flagSet.Set("user", "another test user")
	require.NoError(t, err)
	err = flagSet.Set("db", expectedDB) // Should take precedence over environment
	require.NoError(t, err)
	aDifferentDB := "another test db"
	assert.NotEqual(t, expectedDB, aDifferentDB)
	err = os.Setenv("TEST_DUMMY_CONFIG_DB", aDifferentDB)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_DB", aDifferentDB)
	require.NoError(t, err)
	err = flagSet.Set("int", fmt.Sprintf("%v", expectedInt))
	require.NoError(t, err)
	err = flagSet.Set("time", expectedDuration.String())
	require.NoError(t, err)
	err = flagSet.Set("flag", fmt.Sprintf("%v", false))
	require.NoError(t, err)
	flag, err := flagSet.GetBool("flag")
	require.NoError(t, err)
	assert.False(t, flag)
	assert.False(t, session.GetBool("dummy.config.flag"))
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, expectedString, configTest.TestString)
	assert.Equal(t, expectedInt, configTest.TestInt)
	assert.Equal(t, expectedDuration, configTest.TestTime)
	assert.Equal(t, defaults.TestConfig.Port, configTest.TestConfig.Port)
	assert.Equal(t, expectedHost, configTest.TestConfig.Host)
	assert.Equal(t, expectedHost, configTest.TestConfig2.Host)
	assert.Equal(t, expectedPassword, configTest.TestConfig.Password)
	assert.Equal(t, expectedPassword, configTest.TestConfig2.Password)
	assert.Equal(t, expectedDB, configTest.TestConfig.DB)
	assert.Equal(t, aDifferentDB, configTest.TestConfig2.DB)
	assert.NotEqual(t, expectedDB, configTest.TestConfig2.DB)
	assert.True(t, configTest.TestConfig.Flag)
	assert.False(t, configTest.TestConfig2.Flag)
}

func TestFlagsBinding(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	session := viper.New()
	var err error
	flagSet := pflag.FlagSet{}
	prefix := "test"
	flagSet.String("host1", "a host", "dummy host")
	flagSet.String("host2", "a host", "dummy host")
	flagSet.String("password1", "a password1", "dummy password1")
	flagSet.String("password2", "a password2", "dummy password2")
	flagSet.String("password3", "a password3", "dummy password3")
	flagSet.String("user", "a user", "dummy user")
	flagSet.String("user1", "a user", "dummy user 1")
	flagSet.String("user2", "a user", "dummy user 2")
	flagSet.String("db", "a db", "dummy db")
	flagSet.String("db2", "a db", "dummy db")
	flagSet.Int("int", 0, "dummy int")
	flagSet.Duration("time", time.Second, "dummy time")
	flagSet.Bool("flag", false, "dummy flag")
	err = BindFlagsToEnv(session, prefix, "TEST_DUMMYCONFIG_DUMMY_HOST", flagSet.Lookup("host2"), flagSet.Lookup("host2"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "TEST_DUMMY_CONFIG_DUMMY_HOST", flagSet.Lookup("host1"), flagSet.Lookup("host2"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMYCONFIG_PASSWORD", flagSet.Lookup("password2"), flagSet.Lookup("password3"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMY_CONFIG_PASSWORD", flagSet.Lookup("password1"), flagSet.Lookup("password2"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMYCONFIG_USER", flagSet.Lookup("user"), flagSet.Lookup("user1"), flagSet.Lookup("user2"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMY_CONFIG_USER", flagSet.Lookup("user1"), flagSet.Lookup("user2"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "TEST_DUMMYCONFIG_DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMY_CONFIG_DB", flagSet.Lookup("db2"), flagSet.Lookup("db2"), flagSet.Lookup("db2"), flagSet.Lookup("db2"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMY_CONFIG_FLAG", flagSet.Lookup("flag"), flagSet.Lookup("flag"), flagSet.Lookup("flag"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMY_INT", flagSet.Lookup("int"), flagSet.Lookup("int"), flagSet.Lookup("int"), flagSet.Lookup("int"))
	require.NoError(t, err)
	err = BindFlagsToEnv(session, prefix, "DUMMY_Time", flagSet.Lookup("time"), flagSet.Lookup("time"), flagSet.Lookup("time"))
	require.NoError(t, err)
	err = flagSet.Set("host2", expectedHost)
	require.NoError(t, err)
	fvalue, err := flagSet.GetString("host2")
	require.NoError(t, err)
	assert.Equal(t, expectedHost, fvalue)
	fvalue, err = flagSet.GetString("host1")
	require.NoError(t, err)
	assert.NotEqual(t, expectedHost, fvalue)
	err = flagSet.Set("password2", expectedPassword)
	require.NoError(t, err)
	user1V := faker.Name()
	err = flagSet.Set("user1", user1V)
	require.NoError(t, err)
	user2V := faker.Name()
	err = flagSet.Set("user2", user2V)
	require.NoError(t, err)
	assert.NotEqual(t, user1V, user2V)
	err = flagSet.Set("db", expectedDB) // Should take precedence over environment
	require.NoError(t, err)
	aDifferentDB := "another test db"
	assert.NotEqual(t, expectedDB, aDifferentDB)
	err = os.Setenv("TEST_DUMMY_CONFIG_DB", aDifferentDB)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_DB", aDifferentDB)
	require.NoError(t, err)
	err = flagSet.Set("int", fmt.Sprintf("%v", expectedInt))
	require.NoError(t, err)
	err = flagSet.Set("time", expectedDuration.String())
	require.NoError(t, err)
	err = flagSet.Set("flag", fmt.Sprintf("%v", false))
	require.NoError(t, err)
	flag, err := flagSet.GetBool("flag")
	require.NoError(t, err)
	assert.False(t, flag)
	assert.False(t, session.GetBool("dummy.config.flag"))
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, expectedString, configTest.TestString)
	assert.Equal(t, expectedInt, configTest.TestInt)
	assert.Equal(t, expectedDuration, configTest.TestTime)
	assert.Equal(t, defaults.TestConfig.Port, configTest.TestConfig.Port)
	assert.Contains(t, []string{user1V, user2V}, configTest.TestConfig2.User)
	assert.Equal(t, expectedHost, configTest.TestConfig.Host)
	assert.Equal(t, expectedHost, configTest.TestConfig2.Host)
	assert.Equal(t, expectedPassword, configTest.TestConfig.Password)
	assert.Equal(t, expectedPassword, configTest.TestConfig2.Password)
	assert.Equal(t, expectedDB, configTest.TestConfig.DB)
	assert.Equal(t, aDifferentDB, configTest.TestConfig2.DB)
	assert.NotEqual(t, expectedDB, configTest.TestConfig2.DB)
	assert.True(t, configTest.TestConfig.Flag)
	assert.False(t, configTest.TestConfig2.Flag)
}

func TestFlagsBindingErrors(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	flagSet := pflag.FlagSet{}
	prefix := "test"
	flagSet.String("db2", "a db", "dummy db")
	flagSet.Int("int", 0, "dummy int")
	err := BindFlagsToEnv(session, prefix, "TEST_DUMMYCONFIG_DUMMY_HOST")
	errortest.AssertError(t, err, commonerrors.ErrUndefined)
	err = BindFlagsToEnv(session, prefix, "TEST_DUMMYCONFIG_DUMMY_HOST", flagSet.Lookup("db2"), flagSet.Lookup("int"))
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

}

func TestFlagBindingDefaults(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	session := viper.New()
	var err error
	flagSet := pflag.FlagSet{}
	prefix := "test"
	anotherHostName := fmt.Sprintf("another host %v", faker.DomainName())
	flagSet.String("host", expectedHost, "dummy host")
	flagSet.String("host2", anotherHostName, "dummy host")
	flagSet.String("password", expectedPassword, "dummy password")
	flagSet.String("user", "a user", "dummy user")
	aDifferentDB := "A different db"
	assert.NotEqual(t, expectedDB, aDifferentDB)
	flagSet.String("db", aDifferentDB, "dummy db")
	flagSet.Int("int", expectedInt, "dummy int")
	flagSet.Duration("time", expectedDuration, "dummy time")
	flagSet.Bool("flag", !DefaultDummyConfiguration().Flag, "dummy flag")
	err = BindFlagToEnv(session, prefix, "TEST_DUMMYCONFIG_DUMMY_HOST", flagSet.Lookup("host"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "TEST_DUMMY_CONFIG_DUMMY_HOST", flagSet.Lookup("host2"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMYCONFIG_PASSWORD", flagSet.Lookup("password"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_PASSWORD", flagSet.Lookup("password"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMYCONFIG_USER", flagSet.Lookup("user"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_USER", flagSet.Lookup("user"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "TEST_DUMMYCONFIG_DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_FLAG", flagSet.Lookup("flag"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_INT", flagSet.Lookup("int"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_Time", flagSet.Lookup("time"))
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_DB", expectedDB) // Should take precedence over flag default
	require.NoError(t, err)
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, expectedString, configTest.TestString)
	assert.Equal(t, expectedInt, configTest.TestInt)
	// Defaults from the default structure provided take precedence over defaults from flags when not empty.
	assert.NotEqual(t, expectedDuration, configTest.TestTime)
	assert.Equal(t, DefaultConfiguration().TestTime, configTest.TestTime)
	assert.Equal(t, defaults.TestConfig.Port, configTest.TestConfig.Port)
	assert.NotEqual(t, anotherHostName, expectedHost)
	assert.Equal(t, expectedHost, configTest.TestConfig.Host)
	assert.Equal(t, anotherHostName, configTest.TestConfig2.Host)
	assert.Equal(t, expectedPassword, configTest.TestConfig.Password)
	assert.Equal(t, expectedPassword, configTest.TestConfig2.Password)
	assert.Equal(t, aDifferentDB, configTest.TestConfig.DB)
	assert.Equal(t, expectedDB, configTest.TestConfig2.DB)
	// Defaults from the default structure provided take precedence over defaults from flags when empty.
	assert.Equal(t, DefaultConfiguration().TestConfig.Flag, configTest.TestConfig.Flag)
	assert.Equal(t, DefaultConfiguration().TestConfig.Flag, configTest.TestConfig2.Flag)
}

// Test you can use a struct to load the default env vars
func TestGenerateEnvFile_Defaults(t *testing.T) {
	configTest := DefaultDummyConfiguration()
	prefix := "test"

	// Create test data
	testValues := map[string]interface{}{
		"TEST_DB":                 configTest.DB,
		"TEST_DUMMY_HOST":         configTest.Host,
		"TEST_FLAG":               configTest.Flag,
		"TEST_HEALTHCHECK_PERIOD": configTest.HealthCheckPeriod,
		"TEST_PASSWORD":           configTest.Password,
		"TEST_PORT":               configTest.Port,
		"TEST_USER":               configTest.User,
	}

	// Generate env file
	vars, err := DetermineConfigurationEnvironmentVariables(prefix, configTest)
	require.NoError(t, err)

	// Go through generated vars and check they match the defaults
	for key, value := range vars {
		require.Equal(t, value, testValues[key])
	}
}

// Test that you can load the struct with viper then generate the env file
func TestGenerateEnvFile_Populated(t *testing.T) {
	// Load configuartion using viper
	os.Clearenv()
	configTest := &DummyConfiguration{}
	defaults := DefaultDummyConfiguration()
	session := viper.New()
	var err error
	flagSet := pflag.FlagSet{}
	prefix := "test"
	flagSet.String("host", "a host", "dummy host")
	flagSet.String("password", "a password", "dummy password")
	flagSet.String("user", "a user", "dummy user")
	flagSet.String("db", "a db", "dummy db")
	err = BindFlagToEnv(session, prefix, "TEST_DUMMY_HOST", flagSet.Lookup("host"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "PASSWORD", flagSet.Lookup("password"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "TEST_DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DB", flagSet.Lookup("db"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "USER", flagSet.Lookup("user"))
	require.NoError(t, err)
	err = flagSet.Set("host", expectedHost)
	require.NoError(t, err)
	err = flagSet.Set("password", expectedPassword)
	require.NoError(t, err)
	err = flagSet.Set("db", expectedDB)
	require.NoError(t, err)
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())

	// Create test data
	testValues := map[string]interface{}{
		"TEST_DB":                 configTest.DB,
		"TEST_DUMMY_HOST":         configTest.Host,
		"TEST_FLAG":               configTest.Flag,
		"TEST_HEALTHCHECK_PERIOD": configTest.HealthCheckPeriod,
		"TEST_PASSWORD":           configTest.Password,
		"TEST_PORT":               configTest.Port,
		"TEST_USER":               configTest.User,
	}

	// Generate env file
	vars, err := DetermineConfigurationEnvironmentVariables(prefix, configTest)
	require.NoError(t, err)

	// Go through generated vars and check they match the defaults
	for key, value := range vars {
		require.Equal(t, value, testValues[key])
	}
}

func TestGenerateEnvFile_Nested(t *testing.T) {
	configTest := DefaultDeepConfiguration()
	prefix := "test"

	// Deep nested test values
	/*
		type ConfigurationTest struct {
			TestString  string             `mapstructure:"dummy_string"`
			TestInt     int                `mapstructure:"dummy_int"`
			TestTime    time.Duration      `mapstructure:"dummy_time"`
			TestConfig  DummyConfiguration `mapstructure:"dummyconfig"`  <- nested
			TestConfig2 DummyConfiguration `mapstructure:"dummy_config"` <- nested
		}

		type DeepConfig struct {
			TestString     string            `mapstructure:"dummy_string"`
			TestConfigDeep ConfigurationTest `mapstructure:"deep_config"` <- nested
		}

	*/
	testValues := map[string]interface{}{
		"TEST_DEEP_CONFIG_DUMMYCONFIG_DB":                  configTest.TestConfigDeep.TestConfig.DB,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_DUMMY_HOST":          configTest.TestConfigDeep.TestConfig.Host,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_FLAG":                configTest.TestConfigDeep.TestConfig.Flag,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_HEALTHCHECK_PERIOD":  configTest.TestConfigDeep.TestConfig.HealthCheckPeriod,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_PASSWORD":            configTest.TestConfigDeep.TestConfig.Password,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_PORT":                configTest.TestConfigDeep.TestConfig.Port,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_USER":                configTest.TestConfigDeep.TestConfig.User,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_DB":                 configTest.TestConfigDeep.TestConfig2.DB,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_DUMMY_HOST":         configTest.TestConfigDeep.TestConfig2.Host,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_FLAG":               configTest.TestConfigDeep.TestConfig2.Flag,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_HEALTHCHECK_PERIOD": configTest.TestConfigDeep.TestConfig2.HealthCheckPeriod,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_PASSWORD":           configTest.TestConfigDeep.TestConfig2.Password,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_PORT":               configTest.TestConfigDeep.TestConfig2.Port,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_USER":               configTest.TestConfigDeep.TestConfig2.User,
		"TEST_DEEP_CONFIG_DUMMY_INT":                       configTest.TestConfigDeep.TestInt,
		"TEST_DUMMY_STRING":                                configTest.TestString,
		"TEST_DEEP_CONFIG_DUMMY_TIME":                      configTest.TestConfigDeep.TestTime,
		"TEST_DEEP_CONFIG_DUMMY_STRING":                    configTest.TestConfigDeep.TestString,
	}

	// Generate env file
	vars, err := DetermineConfigurationEnvironmentVariables(prefix, configTest)
	require.NoError(t, err)

	// Go through generated vars and check they match the defaults
	for key, value := range vars {
		require.Equal(t, value, testValues[key])
	}
}

func TestGenerateEnvFile_Undefined(t *testing.T) {
	prefix := "test"
	_, err := DetermineConfigurationEnvironmentVariables(prefix, nil)
	require.ErrorIs(t, err, commonerrors.ErrUndefined)
}

func TestGenerateEnvFile_Empty(t *testing.T) {
	prefix := "test"
	_, err := DetermineConfigurationEnvironmentVariables(prefix, struct{ IServiceConfiguration }{})
	require.ErrorIs(t, err, commonerrors.ErrUndefined)
}

func Test_convertViperError(t *testing.T) {
	tests := []struct {
		viperErr      error
		expectedError error
	}{
		{
			viperErr:      nil,
			expectedError: nil,
		},
		{
			viperErr:      viper.ConfigFileNotFoundError{},
			expectedError: commonerrors.ErrNotFound,
		},
		// Note: the following errors were considered but could not be created outside the viper module (non exposed fields)
		// {
		//	viperErr:      viper.ConfigParseError{},
		//	expectedError: commonerrors.ErrMarshalling,
		// },
		// {
		//	viperErr:      viper.ConfigMarshalError{},
		//	expectedError: commonerrors.ErrMarshalling,
		// },
		{
			viperErr:      viper.UnsupportedConfigError(faker.Sentence()),
			expectedError: commonerrors.ErrUnsupported,
		},
		{
			viperErr:      viper.UnsupportedRemoteProviderError(faker.Sentence()),
			expectedError: commonerrors.ErrUnsupported,
		},
		{
			viperErr:      viper.ConfigFileAlreadyExistsError(faker.Sentence()),
			expectedError: commonerrors.ErrUnexpected,
		},
		{
			viperErr:      viper.RemoteConfigError(faker.Sentence()),
			expectedError: commonerrors.ErrUnexpected,
		},
		{
			viperErr:      errors.New(faker.Name()),
			expectedError: commonerrors.ErrUnexpected,
		},
	}
	for i := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			test := tests[i]
			errortest.RequireError(t, convertViperError(test.viperErr), test.expectedError)
		})
	}
}

func TestServiceConfigurationLoadFromFile(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	err := LoadFromConfigurationFile(session, "")
	assert.Error(t, err)
	err = LoadFromConfigurationFile(session, fmt.Sprintf("doesnotexist-%v.test", faker.DomainName()))
	assert.Error(t, err)
	err = LoadFromConfigurationFile(session, filepath.Join(".", "fixtures", "config-test.json"))
	assert.NoError(t, err)
	value := session.Get("dummy_string")
	assert.NotEmpty(t, value)
	assert.Equal(t, "test string", value)
}

type testCfg struct {
	Field1 string `mapstructure:"f1"`
	Field2 string `mapstructure:"dummy_string"`
}

func (cfg *testCfg) Validate() error {
	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Field1, validation.Required),
	)
}

func TestServiceConfigurationLoadFromFileWithDefaults(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	test := testCfg{}
	err := LoadFromEnvironment(session, "", &test, &testCfg{
		Field1: "test",
	}, filepath.Join(".", "fixtures", "config-test.json"))
	require.NoError(t, err)
	assert.NotEmpty(t, test.Field1)
	assert.Equal(t, "test", test.Field1)
	assert.NotEmpty(t, test.Field2)
	assert.Equal(t, "test string", test.Field2)
}

func TestServiceConfigurationLoadFromEnvironment(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	err := LoadFromEnvironment(session, "test", configTest, defaults, filepath.Join(".", "fixtures", "config-test.json"))
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, "test string", configTest.TestString)
	assert.Equal(t, 1, configTest.TestInt)
	assert.Equal(t, 54*time.Second, configTest.TestTime)
	assert.Equal(t, 20, configTest.TestConfig.Port)
	assert.Equal(t, "host1", configTest.TestConfig.Host)
	assert.Equal(t, "password2", configTest.TestConfig2.Password)
	assert.Equal(t, "db2", configTest.TestConfig2.DB)
	assert.NotEqual(t, expectedHost, configTest.TestConfig2.Host)
	assert.NotEqual(t, expectedPassword, configTest.TestConfig.Password)
	assert.NotEqual(t, expectedDB, configTest.TestConfig.DB)
	assert.True(t, configTest.TestConfig.Flag)
	assert.False(t, configTest.TestConfig2.Flag)
}
