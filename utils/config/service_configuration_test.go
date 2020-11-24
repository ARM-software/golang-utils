package config

import (
	"math/rand"
	"os"
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

type DummyConfiguration struct {
	Host string `mapstructure:"host"`
	Port int `mapstructure:"port"`
	DB string `mapstructure:"db"`
	User string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	HealthCheckPeriod time.Duration `mapstructure:"healthcheck_period"`
}

func (cfg *DummyConfiguration) Validate() error {
	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.DB, validation.Required),
		validation.Field(&cfg.User, validation.Required),
		validation.Field(&cfg.Password, validation.Required),
	)
}

func DefaultDummyConfiguration() DummyConfiguration {
	return DummyConfiguration{
		Port:              5432,
		HealthCheckPeriod: time.Second,
	}
}

var (
	expectedString   = "a test string"
	expectedInt      = rand.Int()
	expectedDuration = time.Hour
	expectedHost     = "a test host"
)

type ConfigurationTest struct {
	TestString string                `mapstructure:"build_artefact_dir"`
	TestInt    int                   `mapstructure:"cmsis_build_default_timeout"` // How long a build is allowed to queue
	TestTime   time.Duration         `mapstructure:"default_build_ttl"`           // How long a build is kept around for
	TestConfig DummyConfiguration 	 `mapstructure:"dummyconfig"`
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
		TestString: expectedString,
		TestInt:    expectedInt,
		TestTime:   expectedDuration,
		TestConfig: DefaultDummyConfiguration(),
	}
}

func TestServiceConfigurationLoad(t *testing.T) {
	configTest := &ConfigurationTest{}
	defaults := DefaultConfiguration()
	err := Load("test", configTest, defaults)
	// Some required values are missing.
	require.NotNil(t, err)
	require.NotNil(t, configTest.Validate())
	// Setting required entries in the environment.
	err = os.Setenv("TEST_DUMMYCONFIG_HOST", expectedHost)
	require.Nil(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_PASSWORD", "a test password")
	require.Nil(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_USER", "a test user")
	require.Nil(t, err)
	err = os.Setenv("TEST_DUMMYCONFIG_DB", "a test db")
	require.Nil(t, err)
	err = Load("test", configTest, defaults)
	require.Nil(t, err)
	require.Nil(t, configTest.Validate())
	assert.Equal(t, configTest.TestString, expectedString)
	assert.Equal(t, configTest.TestInt, expectedInt)
	assert.Equal(t, configTest.TestTime, expectedDuration)
	assert.Equal(t, configTest.TestConfig.Port, defaults.TestConfig.Port)
	assert.Equal(t, configTest.TestConfig.Host, expectedHost)
}
