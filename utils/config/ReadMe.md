# CLI / Service configuration

[Viper](https://github.com/spf13/viper) provides quite a lot of utilities to retrieve configuration values and integrates well with [cobra](https://github.com/spf13/cobra) which is a reference for CLI implementation in go.

Nonetheless `viper` has some gaps we tried to fill for ease of configuration:
- configuration value validation at configuration load
- `.env` configuration
- easy mapping/deserialisation between environment variables and complex nested configuration structures.

The idea is to have the ability to have a complex configuration structure for the project so that values can easily be shared throughout it without requiring viper everywhere. We also tried to gather value in a tree structure in order to categorise them into components configuration instead of having all configuration at the same level in a key/value store. 

## Usage
# Configuration structure
The first step is to define a complex configuration structure for your project and add a `Validate()` method at each level which will be called during load in order to ensure values are as expected.
For validation, we use `github.com/go-ozzo/ozzo-validation/v4` but other libraries/method could be used. It is also advised to provide a method which returns the structure with defaults.

Please look at the following as an example:

```go
type DummyConfiguration struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	DB                string        `mapstructure:"db"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	HealthCheckPeriod time.Duration `mapstructure:"healthcheck_period"`
}

func (cfg *DummyConfiguration) Validate() error {
	// Validate Embedded Structs
	err := ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

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
```
Note the `mapstructure` tag. This will be used as part of the environment value name.
The structure can be complex and of different levels. The name of the environment value mapped to a configuration field will be the combination of the mapstructure tags separated by underscore `_`.
for example, `CLI_HEALTHCHECK_PERIOD` environment variable will refer to the `HealthCheckPeriod` field in the configuration structure above if the prefix used when loading configuration (see below) is `CLI`.


## Configuration load
`Load` or `LoadFromViper` will load values from the environment (including from `.env` files) and assign them to the different configuration fields in the structure passed. Type conversion will be done automatically and if not set and a default value is provided, then the default value will be retained.
Configuration values are then validated using the different `Validate()` methods provided.
In order to easily identify environment variables, a prefix can be provided when loading the configuration. This will tell the system to only consider environment variables with this prefix.

## Integration with Cobra CLI arguments
`cobra` and `viper` are well integrated in the way that you can bind CLI arguments with configuration values.
We provide some utilities (`BindFlagToEnv`) to leverage this binding so that it works with the configuration system described above and limits the number of hardcoded names or global values usually seen in CLI code generated from cobra.
```go
    // Create a viper session instead of using global configuration
	session := viper.New()

	config := &ConfigurationStructConTainingFlag1{}
	defaults := DefaultConfiguration()

	flagSet := pflag.FlagSet{}
	prefix := "ENV_PREFIX"
    // Define CLI flags
	flagSet.String("f", "flag", "a cli flag")
    // Bind flags to environment variables
	err = BindFlagToEnv(session, prefix, "ENV_PREFIX_FLAG1", flagSet.Lookup("flag"))
    ...
    // Load configuration from the environment
    err = LoadFromViper(session, prefix, config, defaults)
```




