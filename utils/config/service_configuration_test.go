/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/keyring"
	mapstest "github.com/ARM-software/golang-utils/utils/serialization/maps/testing" //nolint:misspell
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
	Host              string                            `mapstructure:"dummy_host"`
	Port              int                               `mapstructure:"port"`
	DB                string                            `mapstructure:"db"`
	User              string                            `mapstructure:"user"`
	Password          string                            `mapstructure:"password"`
	Flag              bool                              `mapstructure:"flag"`
	TestEnum          mapstest.TestEnumWithUnmarshal    `mapstructure:"enum"`
	TestEnum1         mapstest.TestEnumWithoutUnmarshal `mapstructure:"enum1"`
	HealthCheckPeriod time.Duration                     `mapstructure:"healthcheck_period"`
}

func (cfg *DummyConfiguration) Validate() error {
	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required, validation.Min(0)),
		validation.Field(&cfg.DB, validation.Required),
		validation.Field(&cfg.User, validation.Required),
		validation.Field(&cfg.Password, validation.Required),
		validation.Field(&cfg.TestEnum, validation.By(mapstest.ValidationFunc)),
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
	// Validate Embedded Structs
	err := ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.TestConfigDeep, validation.Required),
	)
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

func TestErrorFormatting(t *testing.T) {
	cfg := DefaultConfiguration()
	err := cfg.Validate()
	require.Error(t, err)

	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Contains(t, err.Error(), "invalid: structure failed validation: (TestConfig->db) [DUMMYCONFIG_DB] cannot be blank")
}

func TestDeepErrorFormatting(t *testing.T) {
	defaults := DefaultDeepConfiguration()
	err := defaults.Validate()
	require.Error(t, err)

	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Contains(t, err.Error(), "invalid: structure failed validation: (TestConfigDeep->TestConfig->db) [DEEP_CONFIG_DUMMYCONFIG_DB] cannot be blank")

	err = os.Setenv("TEST_DEEP_CONFIG_DUMMYCONFIG_DB", "a test db")
	require.NoError(t, err)
	err = os.Setenv("TEST_DEEP_CONFIG_DUMMYCONFIG_DUMMY_HOST", "a test host")
	require.NoError(t, err)
	err = os.Setenv("TEST_DEEP_CONFIG_DUMMYCONFIG_PASSWORD", "a test password")
	require.NoError(t, err)
	err = os.Setenv("TEST_DEEP_CONFIG_DUMMYCONFIG_USER", "a test user")
	require.NoError(t, err)
	err = os.Setenv("TEST_DEEP_CONFIG_DUMMY_CONFIG_DB", "a test user")
	require.NoError(t, err)

	t.Run("defined mapstructure", func(t *testing.T) {
		configTest2 := &DeepConfig{}
		err = LoadFromSystem("test", configTest2, defaults)

		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Contains(t, err.Error(), "invalid: structure failed validation: (TestConfigDeep->TestConfig2->dummy_host) [TEST_DEEP_CONFIG_DUMMY_CONFIG_DUMMY_HOST] cannot be blank")
	})
}

func TestServiceConfigurationLoad(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	require.NoError(t, keyring.Clear(context.Background(), "test"))
	err := Load("test", configTest, defaults)
	// Some required values are missing.
	require.Error(t, err)

	assert.ErrorContains(t, err, "(TestConfig->db) [TEST_DUMMYCONFIG_DB] cannot be blank")

	errortest.RequireError(t, err, commonerrors.ErrInvalid)
	errortest.RequireError(t, configTest.Validate(), commonerrors.ErrInvalid)

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
	err = os.Setenv("TEST_DUMMY_CONFIG_ENUM", mapstest.TestEnumStringVer1)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_ENUM", mapstest.TestEnumStringVer1)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_ENUM1", "1")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_ENUM1", "1")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_DB", "a test db")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_DB", expectedDB)
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_CONFIG_FLAG", "false")
	require.NoError(t, err)
	err = os.Setenv("TEST_DUMMY_TIME", expectedDuration.String())
	require.NoError(t, err)
	err = Load("test", configTest, defaults)
	errortest.RequireError(t, err, commonerrors.ErrInvalid)
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
	t.Run("load from system", func(t *testing.T) {
		configTest2 := &ConfigurationTest{}
		err = LoadFromSystem("test", configTest2, defaults)
		require.NoError(t, err)
		require.NoError(t, configTest2.Validate())
	})
	t.Run("load from system", func(t *testing.T) {
		configTest2 := &ConfigurationTest{}
		err = Load("test", configTest2, defaults)
		require.NoError(t, err)
		require.NoError(t, configTest2.Validate())
		assert.EqualExportedValues(t, configTest, configTest2)
		configTest2.TestConfig2.Host = faker.URL()
		configTest2.TestConfig2.User = faker.Name()
		assert.NotEqual(t, configTest, configTest2)
		err := keyring.Store[ConfigurationTest](context.Background(), "test", configTest2)
		errortest.AssertError(t, err, nil, commonerrors.ErrUnsupported)
		if commonerrors.Any(err, commonerrors.ErrUnsupported) {
			t.Skip("keyring is not supported")
		}
		configTest3 := &ConfigurationTest{}
		err = LoadFromSystem("test", configTest3, defaults)
		require.NoError(t, err)
		require.NoError(t, configTest3.Validate())
		assert.EqualExportedValues(t, configTest2, configTest3)
		assert.NotEqual(t, configTest, configTest3)
		configTest4 := &ConfigurationTest{}
		err = Load("test", configTest4, defaults)
		require.NoError(t, err)
		require.NoError(t, configTest4.Validate())
		assert.EqualExportedValues(t, configTest, configTest4)
		assert.NotEqual(t, configTest4, configTest3)
	})
}

func TestServiceConfigurationLoad_Errors(t *testing.T) {
	os.Clearenv()
	configTest := &ConfigurationTest{}
	err := Load("test", configTest, DefaultConfiguration())
	// Some required values are missing.
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	errortest.AssertError(t, configTest.Validate(), commonerrors.ErrInvalid)

	err = Load("test", nil, DefaultDummyConfiguration())
	// Incorrect  structure provided.
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrMarshalling)

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
	flagSet.String("enum", mapstest.TestEnumStringVer1, "dummy enum")
	flagSet.String("enum1", "1", "dummy enum")
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
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_ENUM", flagSet.Lookup("enum"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_CONFIG_ENUM1", flagSet.Lookup("enum1"))
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
	assert.NotEqual(t, mapstest.TestEnumStringVer1, configTest.TestConfig2.TestEnum)
	assert.NotEqual(t, mapstest.TestEnumStringVer0, configTest.TestConfig.TestEnum)
	assert.Equal(t, mapstest.TestEnumWithUnmarshal1, configTest.TestConfig2.TestEnum)
	assert.Equal(t, mapstest.TestEnumWithUnmarshal0, configTest.TestConfig.TestEnum)
	assert.Equal(t, mapstest.TestEnumWithoutUnmarshal1, configTest.TestConfig2.TestEnum1)
	assert.Equal(t, mapstest.TestEnumWithoutUnmarshal0, configTest.TestConfig.TestEnum1)
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
	flagSet.String("enum", mapstest.TestEnumStringVer0, "dummy enum")
	flagSet.String("enum1", "0", "dummy enum")
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
	err = BindFlagToEnv(session, prefix, "DUMMY_enum", flagSet.Lookup("enum"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "DUMMY_enum1", flagSet.Lookup("enum1"))
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
	assert.Equal(t, mapstest.TestEnumWithUnmarshal0, configTest.TestConfig2.TestEnum)
	assert.Equal(t, mapstest.TestEnumWithUnmarshal0, configTest.TestConfig.TestEnum)
	assert.Equal(t, mapstest.TestEnumWithoutUnmarshal0, configTest.TestConfig2.TestEnum1)
	assert.Equal(t, mapstest.TestEnumWithoutUnmarshal0, configTest.TestConfig.TestEnum1)
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
		"TEST_ENUM":               configTest.TestEnum,
		"TEST_ENUM1":              configTest.TestEnum1,
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
	// Load configuration using viper
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
	flagSet.String("enum", mapstest.TestEnumStringVer1, "dummy enum")
	flagSet.String("enum1", "1", "dummy enum")
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
	err = BindFlagToEnv(session, prefix, "ENUM", flagSet.Lookup("enum"))
	require.NoError(t, err)
	err = BindFlagToEnv(session, prefix, "ENUM1", flagSet.Lookup("enum1"))
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
		"TEST_ENUM":               configTest.TestEnum,
		"TEST_ENUM1":              configTest.TestEnum1,
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
		"TEST_DEEP_CONFIG_DUMMYCONFIG_ENUM":                configTest.TestConfigDeep.TestConfig.TestEnum,
		"TEST_DEEP_CONFIG_DUMMYCONFIG_ENUM1":               configTest.TestConfigDeep.TestConfig.TestEnum1,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_DB":                 configTest.TestConfigDeep.TestConfig2.DB,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_DUMMY_HOST":         configTest.TestConfigDeep.TestConfig2.Host,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_FLAG":               configTest.TestConfigDeep.TestConfig2.Flag,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_HEALTHCHECK_PERIOD": configTest.TestConfigDeep.TestConfig2.HealthCheckPeriod,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_PASSWORD":           configTest.TestConfigDeep.TestConfig2.Password,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_PORT":               configTest.TestConfigDeep.TestConfig2.Port,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_USER":               configTest.TestConfigDeep.TestConfig2.User,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_ENUM":               configTest.TestConfigDeep.TestConfig2.TestEnum,
		"TEST_DEEP_CONFIG_DUMMY_CONFIG_ENUM1":              configTest.TestConfigDeep.TestConfig2.TestEnum1,
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
	errortest.RequireError(t, err, commonerrors.ErrUndefined)
}

func TestGenerateEnvFile_Empty(t *testing.T) {
	prefix := "test"
	_, err := DetermineConfigurationEnvironmentVariables(prefix, struct{ IServiceConfiguration }{})
	errortest.RequireError(t, err, commonerrors.ErrUndefined)
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
	errortest.AssertError(t, err, commonerrors.ErrUndefined, commonerrors.ErrNotFound)
	err = LoadFromConfigurationFile(session, fmt.Sprintf("doesnotexist-%v.test", faker.DomainName()))
	assert.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrUndefined, commonerrors.ErrNotFound, commonerrors.ErrUnsupported)
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

type TestBaseCfg struct {
	Embedded1 string `mapstructure:"embedded1"`
	Embedded2 string `mapstructure:"embedded2"`
}

func (cfg *TestBaseCfg) Validate() error {
	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Embedded1, validation.Required),
		validation.Field(&cfg.Embedded2, validation.Required),
	)
}

type TestCfgWithEmbeddedCfg struct {
	TestBaseCfg
	NonEmbedded1 string `mapstructure:"non_embedded1"`
	NonEmbedded2 string `mapstructure:"non_embedded2"`
}

func (cfg *TestCfgWithEmbeddedCfg) Validate() error {
	// Validate Embedded Structs
	err := ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.NonEmbedded1, validation.Required),
	)
}

type TestCfgWithEmbeddedCfgWithTag struct {
	TestBaseCfg  `mapstructure:"embedded_struct"`
	NonEmbedded1 string `mapstructure:"non_embedded1"`
}

func (cfg *TestCfgWithEmbeddedCfgWithTag) Validate() error {
	// Validate Embedded Structs
	err := ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.NonEmbedded1, validation.Required),
	)
}

type TestCfgWithEmbeddedCfgWithSquashTag struct {
	TestBaseCfg  `mapstructure:",squash"`
	NonEmbedded1 string `mapstructure:"non_embedded1"`
}

func (cfg *TestCfgWithEmbeddedCfgWithSquashTag) Validate() error {
	// Validate Embedded Structs
	err := ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.NonEmbedded1, validation.Required),
	)
}

// Config values loaded from file should be correctly mapped onto a struct that embeds another struct with no mapstructure tag set
func TestEmbeddedServiceConfigurationWithNoTagLoadFromFile(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbedded := TestCfgWithEmbeddedCfg{}
	err := LoadFromEnvironment(session, "", &testEmbedded,
		&TestCfgWithEmbeddedCfg{}, filepath.Join(".", "fixtures", "nested-config-test.json"))
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbedded.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbedded.NonEmbedded1)
	assert.Empty(t, testEmbedded.NonEmbedded2)
	assert.NotEmpty(t, testEmbedded.Embedded1)
	assert.Equal(t, "embedded 1", testEmbedded.Embedded1)
	assert.NotEmpty(t, testEmbedded.Embedded2)
	assert.Equal(t, "embedded 2", testEmbedded.Embedded2)
}

// Nested config values loaded from file should be correctly mapped onto a struct that embeds another struct with a mapstructure tag set
func TestEmbeddedServiceConfigurationWithTagLoadFromFile(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbeddedWithTag := TestCfgWithEmbeddedCfgWithTag{}
	err := LoadFromEnvironment(session, "", &testEmbeddedWithTag,
		&TestCfgWithEmbeddedCfgWithTag{}, filepath.Join(".", "fixtures", "nested-config-test.json"))
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbeddedWithTag.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbeddedWithTag.NonEmbedded1)
	assert.NotEmpty(t, testEmbeddedWithTag.Embedded1)
	assert.Equal(t, "embedded 1", testEmbeddedWithTag.Embedded1)
	assert.NotEmpty(t, testEmbeddedWithTag.Embedded2)
	assert.Equal(t, "embedded 2", testEmbeddedWithTag.Embedded2)
}

// Flat config values loaded from file should be correctly mapped onto a struct that embeds another struct with the ",squash" mapstructure tag set
func TestEmbeddedServiceConfigurationWithSquashTagLoadFromFile(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbeddedWithSquashTag := TestCfgWithEmbeddedCfgWithSquashTag{}
	err := LoadFromEnvironment(session, "", &testEmbeddedWithSquashTag,
		&TestCfgWithEmbeddedCfgWithSquashTag{}, filepath.Join(".", "fixtures", "flat-config-test.json"))
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbeddedWithSquashTag.NonEmbedded1)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.Embedded1)
	assert.Equal(t, "embedded 1", testEmbeddedWithSquashTag.Embedded1)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.Embedded2)
	assert.Equal(t, "embedded 2", testEmbeddedWithSquashTag.Embedded2)
}

// Nested config values loaded from file should not be correctly mapped onto a struct that embeds another struct with the ",squash" mapstructure tag set
func TestEmbeddedServiceConfigurationWithSquashTagLoadFromNestedFile(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbeddedWithSquashTag := TestCfgWithEmbeddedCfgWithSquashTag{}
	err := LoadFromEnvironment(session, "", &testEmbeddedWithSquashTag,
		&TestCfgWithEmbeddedCfgWithSquashTag{}, filepath.Join(".", "fixtures", "nested-config-test.json"))
	require.Error(t, err)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbeddedWithSquashTag.NonEmbedded1)
	assert.Empty(t, testEmbeddedWithSquashTag.Embedded1)
	assert.NotEqual(t, "embedded 1", testEmbeddedWithSquashTag.Embedded1)
	assert.NotEqual(t, "embedded 2", testEmbeddedWithSquashTag.Embedded2)
}

// Nested config values loaded from file should be correctly mapped onto a struct that embeds another struct and should maintain any defaults not overwritten
func TestEmbeddedServiceConfigurationLoadFromFileWithDefaults(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbedded := TestCfgWithEmbeddedCfg{}
	defaults := &TestCfgWithEmbeddedCfg{
		TestBaseCfg: TestBaseCfg{
			Embedded1: "a",
			Embedded2: "b",
		},
		NonEmbedded1: "c",
		NonEmbedded2: "d",
	}
	err := LoadFromEnvironment(session, "", &testEmbedded,
		defaults, filepath.Join(".", "fixtures", "nested-config-test.json"))
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbedded.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbedded.NonEmbedded1)
	assert.NotEmpty(t, testEmbedded.NonEmbedded2)
	assert.Equal(t, "d", testEmbedded.NonEmbedded2)
	assert.NotEmpty(t, testEmbedded.Embedded1)
	assert.Equal(t, "embedded 1", testEmbedded.Embedded1)
	assert.NotEmpty(t, testEmbedded.Embedded2)
	assert.Equal(t, "embedded 2", testEmbedded.Embedded2)
}

// Config values loaded from env vars should be correctly mapped onto a struct that embeds another struct with no mapstructure tag set
func TestEmbeddedServiceConfigurationWithNoTagLoadFromEnvironment(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbedded := TestCfgWithEmbeddedCfg{}
	err := loadEnvIntoEnvironment(t, filepath.Join(".", "fixtures", "env-test.env"))
	require.NoError(t, err)

	err = LoadFromEnvironment(session, "WITH_NO_TAG", &testEmbedded,
		&TestCfgWithEmbeddedCfg{}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbedded.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbedded.NonEmbedded1)
	assert.NotEmpty(t, testEmbedded.Embedded1)
	assert.Equal(t, "embedded 1", testEmbedded.Embedded1)
	assert.NotEmpty(t, testEmbedded.Embedded2)
	assert.Equal(t, "embedded 2", testEmbedded.Embedded2)
}

// Config values loaded from env vars should be correctly mapped onto a struct that embeds another struct with a mapstructure tag set
func TestEmbeddedServiceConfigurationWithTagLoadFromEnvironment(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbeddedWithTag := TestCfgWithEmbeddedCfgWithTag{}
	err := loadEnvIntoEnvironment(t, filepath.Join(".", "fixtures", "env-test.env"))
	require.NoError(t, err)

	err = LoadFromEnvironment(session, "WITH_TAG", &testEmbeddedWithTag,
		&TestCfgWithEmbeddedCfgWithTag{}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbeddedWithTag.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbeddedWithTag.NonEmbedded1)
	assert.NotEmpty(t, testEmbeddedWithTag.Embedded1)
	assert.Equal(t, "embedded 1", testEmbeddedWithTag.Embedded1)
	assert.NotEmpty(t, testEmbeddedWithTag.Embedded2)
	assert.Equal(t, "embedded 2", testEmbeddedWithTag.Embedded2)
}

// Flat config values loaded from env vars should be correctly mapped onto a struct that embeds another struct with the ",squash" mapstructure tag set
func TestEmbeddedServiceConfigurationWithSquashTagLoadFromEnvironment(t *testing.T) {
	os.Clearenv()
	session := viper.New()
	testEmbeddedWithSquashTag := TestCfgWithEmbeddedCfgWithSquashTag{}
	err := loadEnvIntoEnvironment(t, filepath.Join(".", "fixtures", "env-test.env"))
	require.NoError(t, err)

	err = LoadFromEnvironment(session, "WITH_SQUASH_TAG", &testEmbeddedWithSquashTag,
		&TestCfgWithEmbeddedCfgWithSquashTag{}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.NonEmbedded1)
	assert.Equal(t, "non-embedded 1", testEmbeddedWithSquashTag.NonEmbedded1)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.Embedded1)
	assert.Equal(t, "embedded 1", testEmbeddedWithSquashTag.Embedded1)
	assert.NotEmpty(t, testEmbeddedWithSquashTag.Embedded2)
	assert.Equal(t, "embedded 2", testEmbeddedWithSquashTag.Embedded2)
}

func loadEnvIntoEnvironment(t *testing.T, envPath string) (err error) {
	t.Helper()
	_, err = fs.Stat(os.DirFS("."), envPath)
	require.NoError(t, err)

	err = godotenv.Load(envPath)
	require.NoError(t, err)

	return
}

func TestCustomTypeHook_Success(t *testing.T) {
	t.Cleanup(os.Clearenv)
	os.Clearenv()

	cfg := &ConfigurationTest{}
	defaults := DefaultConfiguration()

	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_DUMMY_HOST", expectedHost))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_PASSWORD", expectedPassword))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_USER", "user"))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_DB", expectedDB))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_DUMMY_HOST", expectedHost))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_PASSWORD", expectedPassword))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_USER", "user"))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_DB", expectedDB))
	require.NoError(t, os.Setenv("TEST_DUMMY_INT", fmt.Sprintf("%v", expectedInt)))
	require.NoError(t, os.Setenv("TEST_DUMMY_TIME", expectedDuration.String()))

	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_ENUM", mapstest.TestEnumStringVer1))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_ENUM1", "1"))

	err := Load("test", cfg, defaults)
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	assert.Equal(t, mapstest.TestEnumWithUnmarshal1, cfg.TestConfig.TestEnum)
	assert.Equal(t, mapstest.TestEnumWithoutUnmarshal1, cfg.TestConfig.TestEnum1)
}

func TestCustomTypeHook_InvalidValue(t *testing.T) {
	t.Cleanup(os.Clearenv)
	os.Clearenv()

	cfg := &ConfigurationTest{}
	defaults := DefaultConfiguration()

	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_DUMMY_HOST", expectedHost))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_PASSWORD", expectedPassword))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_USER", "user"))
	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_DB", expectedDB))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_DUMMY_HOST", expectedHost))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_PASSWORD", expectedPassword))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_USER", "user"))
	require.NoError(t, os.Setenv("TEST_DUMMY_CONFIG_DB", expectedDB))
	require.NoError(t, os.Setenv("TEST_DUMMY_INT", fmt.Sprintf("%v", expectedInt)))
	require.NoError(t, os.Setenv("TEST_DUMMY_TIME", expectedDuration.String()))

	require.NoError(t, os.Setenv("TEST_DUMMYCONFIG_ENUM", "4"))

	err := Load("test", cfg, defaults)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "structure failed validation")
}
