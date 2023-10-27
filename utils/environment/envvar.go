package environment

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/platform"
)

var (
	envvarKeyRegex   = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$") // See [IEEE Std 1003.1-2008 / IEEE POSIX P1003.2/ISO 9945.2](http://www.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_10_02)
	errEnvvarInvalid = validation.NewError("validation_is_environment_variable", "must be a valid Posix environment variable")

	// IsEnvironmentVariableKey defines a validation rule for environment variable keys ([IEEE Std 1003.1-2008 / IEEE POSIX P1003.2/ISO 9945.2](http://www.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_10_02)) for use with github.com/go-ozzo/ozzo-validation
	// TODO use the built-in implementation in `is` package when https://github.com/go-ozzo/ozzo-validation/issues/186 is looked at.
	IsEnvironmentVariableKey = validation.NewStringRuleWithError(isEnvVarKey, errEnvvarInvalid)
)

type EnvVar struct {
	key             string
	value           string
	validationRules []validation.Rule
}

func (e *EnvVar) MarshalText() (text []byte, err error) {
	err = e.Validate()
	text = []byte(e.String())
	return
}

func (e *EnvVar) UnmarshalText(text []byte) error {
	env, err := ParseEnvironmentVariable(string(text))
	if err != nil {
		return err
	}
	e.key = env.GetKey()
	e.value = env.GetValue()
	return nil
}

func (e *EnvVar) Equal(v IEnvironmentVariable) bool {
	if v == nil {
		return e == nil
	}
	if e == nil {
		return false
	}
	return e.GetKey() == v.GetKey() && e.GetValue() == v.GetValue()
}

func (e *EnvVar) GetKey() string {
	return e.key
}

func (e *EnvVar) GetValue() string {
	return e.value
}

func (e *EnvVar) String() string {
	return fmt.Sprintf("%v=%v", e.GetKey(), e.GetValue())
}

func (e *EnvVar) Validate() (err error) {
	err = validation.Validate(e.GetKey(), validation.Required, IsEnvironmentVariableKey)
	if err != nil {
		err = fmt.Errorf("%w: environment variable name `%v` is not valid: %v", commonerrors.ErrInvalid, e.GetKey(), err.Error())
		return
	}
	if len(e.validationRules) > 0 {
		err = validation.Validate(e.GetValue(), e.validationRules...)
		if err != nil {
			err = fmt.Errorf("%w: environment variable `%v` value is not valid: %v", commonerrors.ErrInvalid, e.GetKey(), err.Error())
		}
	}
	return
}

func isEnvVarKey(value string) bool {
	// FIXME remove when supported by the validation tool see https://github.com/go-ozzo/ozzo-validation/issues/186
	return envvarKeyRegex.MatchString(value)
}

// ParseEnvironmentVariable parses an environment variable definition, in the form "key=value".
func ParseEnvironmentVariable(variable string) (IEnvironmentVariable, error) {
	elements := strings.Split(strings.TrimSpace(variable), "=")
	if len(elements) < 2 {
		return nil, fmt.Errorf("%w: invalid environment variable entry as not following key=value", commonerrors.ErrInvalid)
	}
	value := elements[1]
	if len(elements) > 2 {
		var valueElems []string
		for i := 1; i < len(elements); i++ {
			valueElems = append(valueElems, elements[i])
		}
		value = strings.Join(valueElems, "=")
	}
	envvar := NewEnvironmentVariable(elements[0], value)
	return envvar, envvar.Validate()
}

// NewEnvironmentVariable returns an environment variable defined by a key and a value.
func NewEnvironmentVariable(key, value string) IEnvironmentVariable {
	return NewEnvironmentVariableWithValidation(key, value)

}

// CloneEnvironementVariable returns a clone of the environment variable.
func CloneEnvironementVariable(envVar IEnvironmentVariable) IEnvironmentVariable {
	if envVar == nil {
		return nil
	}
	return NewEnvironmentVariable(envVar.GetKey(), envVar.GetValue())
}

// NewEnvironmentVariableWithValidation returns an environment variable defined by a key and a value but with the possibility to define value validation rules.
func NewEnvironmentVariableWithValidation(key, value string, rules ...validation.Rule) IEnvironmentVariable {
	return &EnvVar{
		key:             key,
		value:           value,
		validationRules: rules,
	}
}

// ValidateEnvironmentVariables validates that environment variables are correctly defined in regard to their schema.
func ValidateEnvironmentVariables(vars ...IEnvironmentVariable) error {
	for i := range vars {
		err := vars[i].Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseEnvironmentVariables parses a list of key=value entries such as os.Environ() and returns a list of the corresponding environment variables.
// Any entry failing parsing will be ignored.
func ParseEnvironmentVariables(variables ...string) (envVars []IEnvironmentVariable) {
	for i := range variables {
		envvar, err := ParseEnvironmentVariable(variables[i])
		if err != nil {
			continue
		}
		envVars = append(envVars, envvar)
	}
	return
}

// FindEnvironmentVariable looks for an environment variable in a list. if no environment variable matches, an error is returned
func FindEnvironmentVariable(envvar string, envvars ...IEnvironmentVariable) (IEnvironmentVariable, error) {
	for i := range envvars {
		if envvars[i].GetKey() == envvar {
			return envvars[i], nil
		}
	}
	return nil, fmt.Errorf("%w: environment variable '%v' not set", commonerrors.ErrNotFound, envvar)
}

// ExpandEnvironmentVariables returns a list of environment variables with their value being expanded.
// Expansion assumes that all the variables are present in the envvars list.
// If recursive is set to true, then expansion is performed recursively over the variable list.
func ExpandEnvironmentVariables(recursive bool, envvars ...IEnvironmentVariable) (expandedEnvVars []IEnvironmentVariable) {
	for i := range envvars {
		expandedEnvVars = append(expandedEnvVars, ExpandEnvironmentVariable(recursive, envvars[i], envvars...))
	}
	return
}

// ExpandEnvironmentVariable returns a clone of envVarToExpand but with an expanded value based on environment variables defined in envvars list.
// Expansion assumes that all the variables are present in the envvars list.
// If recursive is set to true, then expansion is performed recursively over the variable list.
func ExpandEnvironmentVariable(recursive bool, envVarToExpand IEnvironmentVariable, envvars ...IEnvironmentVariable) (expandedEnvVar IEnvironmentVariable) {
	if len(envvars) == 0 || envVarToExpand == nil {
		return envVarToExpand
	}
	mappingFunc := func(envvarKey string) (string, bool) {
		envVar, err := FindEnvironmentVariable(envvarKey, envvars...)
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			return "", false
		}
		return envVar.GetValue(), true
	}
	expandedEnvVar = NewEnvironmentVariable(envVarToExpand.GetKey(), platform.ExpandParameter(envVarToExpand.GetValue(), mappingFunc, recursive))
	return
}
