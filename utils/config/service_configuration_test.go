/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expectedString   = fmt.Sprintf("a test string %v", faker.Word())
	expectedInt      = rand.Int() //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	expectedDuration = time.Since(time.Date(1999, 2, 3, 4, 30, 45, 46, time.UTC))
	expectedHost     = fmt.Sprintf("a test host %v", faker.Word())
	expectedPassword = fmt.Sprintf("a test passwd %v", faker.Password())
)

type DummyConfiguration struct {
	Host              string        `mapstructure:"dummy_host"`
	Port              int           `mapstructure:"port"`
	DB                string        `mapstructure:"db"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
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
	err = os.Setenv("TEST_DUMMY_CONFIG_DB", "a test db")
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
	assert.NotEqual(t, expectedHost, configTest.TestConfig2.Host)
	assert.NotEqual(t, expectedPassword, configTest.TestConfig.Password)
	assert.Equal(t, expectedPassword, configTest.TestConfig2.Password)
}

func TestBinding(t *testing.T) {
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
	flagSet.Int("int", 0, "dummy int")
	flagSet.Duration("time", time.Second, "dummy time")
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
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_DB", flagSet.Lookup("db"))
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
	err = flagSet.Set("db", "another test db")
	require.NoError(t, err)
	err = flagSet.Set("int", fmt.Sprintf("%v", expectedInt))
	require.NoError(t, err)
	err = flagSet.Set("time", expectedDuration.String())
	require.NoError(t, err)
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, configTest.TestString, expectedString)
	assert.Equal(t, configTest.TestInt, expectedInt)
	assert.Equal(t, configTest.TestTime, expectedDuration)
	assert.Equal(t, configTest.TestConfig.Port, defaults.TestConfig.Port)
	assert.Equal(t, configTest.TestConfig.Host, expectedHost)
	assert.Equal(t, configTest.TestConfig2.Host, expectedHost)
	assert.Equal(t, configTest.TestConfig.Password, expectedPassword)
	assert.Equal(t, configTest.TestConfig2.Password, expectedPassword)
}

func TestBindingDefaults(t *testing.T) {
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
	flagSet.String("db", "a db", "dummy db")
	flagSet.Int("int", expectedInt, "dummy int")
	flagSet.Duration("time", expectedDuration, "dummy time")
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
	err = BindFlagToEnv(session, prefix, "DUMMY_INT", flagSet.Lookup("int"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_Time", flagSet.Lookup("time"))
	require.NoError(t, err)
	
	err = LoadFromViper(session, prefix, configTest, defaults)
	require.NoError(t, err)
	require.NoError(t, configTest.Validate())
	assert.Equal(t, configTest.TestString, expectedString)
	assert.Equal(t, configTest.TestInt, expectedInt)
	assert.Equal(t, configTest.TestTime, expectedDuration)
	assert.Equal(t, configTest.TestConfig.Port, defaults.TestConfig.Port)
	assert.NotEqual(t, expectedHost, anotherHostName)
	assert.Equal(t, configTest.TestConfig.Host, expectedHost)
	assert.Equal(t, configTest.TestConfig2.Host, anotherHostName)
	assert.Equal(t, configTest.TestConfig.Password, expectedPassword)
	assert.Equal(t, configTest.TestConfig2.Password, expectedPassword)
}
