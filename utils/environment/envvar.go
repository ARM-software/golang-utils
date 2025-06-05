package environment

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/exp/maps"

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
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "environment variable name `%v` is not valid", e.GetKey())
		return
	}
	if len(e.validationRules) > 0 {
		err = validation.Validate(e.GetValue(), e.validationRules...)
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "environment variable `%v` value is not valid", e.GetKey())
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
		return nil, commonerrors.New(commonerrors.ErrInvalid, "invalid environment variable entry as not following key=value")
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

// CloneEnvironmentVariable returns a clone of the environment variable.
func CloneEnvironmentVariable(envVar IEnvironmentVariable) IEnvironmentVariable {
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
	return nil, commonerrors.Newf(commonerrors.ErrNotFound, "environment variable '%v' not set", envvar)
}

// FindEnvironmentVariables looks for environment variables in a list. if no environment variable matches, an error is returned
func FindEnvironmentVariables(environment []IEnvironmentVariable, envvarToSearchFor ...string) ([]IEnvironmentVariable, error) {
	envs := make([]IEnvironmentVariable, 0, len(envvarToSearchFor))
	for i := range envvarToSearchFor {
		found, err := FindEnvironmentVariable(envvarToSearchFor[i], environment...)
		if err == nil && found != nil {
			envs = append(envs, found)
		}
	}
	if len(envs) == 0 {
		return nil, commonerrors.Newf(commonerrors.ErrNotFound, "could not find any environment variables %v set", envvarToSearchFor)
	}
	return envs, nil
}

// FindFoldEnvironmentVariable looks for an environment variable in a list similarly to FindEnvironmentVariable but without case-sensitivity.
func FindFoldEnvironmentVariable(envvar string, envvars ...IEnvironmentVariable) (IEnvironmentVariable, error) {
	for i := range envvars {
		if strings.EqualFold(envvars[i].GetKey(), envvar) {
			return envvars[i], nil
		}
	}
	return nil, commonerrors.Newf(commonerrors.ErrNotFound, "environment variable '%v' not set", envvar)
}

// FindFoldEnvironmentVariables looks for environment variables in a list with no case-sensitivity. if no environment variable matches, an error is returned
func FindFoldEnvironmentVariables(environment []IEnvironmentVariable, envvarToSearchFor ...string) ([]IEnvironmentVariable, error) {
	envs := make([]IEnvironmentVariable, 0, len(envvarToSearchFor))
	for i := range envvarToSearchFor {
		found, err := FindFoldEnvironmentVariable(envvarToSearchFor[i], environment...)
		if err == nil && found != nil {
			envs = append(envs, found)
		}
	}
	if len(envs) == 0 {
		return nil, commonerrors.Newf(commonerrors.ErrNotFound, "could not find any environment variables %v set", envvarToSearchFor)
	}
	return envs, nil
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

// UniqueEnvironmentVariables returns a list of unique environment variables.
// caseSensitive states whether two same keys but with different case should be both considered unique.
func UniqueEnvironmentVariables(caseSensitive bool, envvars ...IEnvironmentVariable) (uniqueEnvVars []IEnvironmentVariable) {
	uniqueSet := map[string]IEnvironmentVariable{}
	recordUniqueEnvVar(caseSensitive, envvars, uniqueSet)
	uniqueEnvVars = maps.Values(uniqueSet)
	return
}

// SortEnvironmentVariables sorts a list of environment variable alphabetically no matter the case.
func SortEnvironmentVariables(envvars []IEnvironmentVariable) {
	if len(envvars) == 0 {
		return
	}
	sort.SliceStable(envvars, func(i, j int) bool {
		return strings.ToLower(envvars[i].GetKey()) < strings.ToLower(envvars[j].GetKey())
	})
}

// MergeEnvironmentVariableSets merges two sets of environment variables.
// If both sets have a same environment variable, its value in set 1 will take precedence.
// caseSensitive states whether two similar keys with different case should be considered as different
func MergeEnvironmentVariableSets(caseSensitive bool, envvarSet1 []IEnvironmentVariable, envvarSet2 ...IEnvironmentVariable) (mergedEnvVars []IEnvironmentVariable) {
	mergeSet := map[string]IEnvironmentVariable{}
	recordUniqueEnvVar(caseSensitive, envvarSet1, mergeSet)
	recordUniqueEnvVar(caseSensitive, envvarSet2, mergeSet)
	mergedEnvVars = maps.Values(mergeSet)
	return
}

func recordUniqueEnvVar(caseSensitive bool, envvarSet []IEnvironmentVariable, hashTable map[string]IEnvironmentVariable) {
	for i := range envvarSet {
		key := envvarSet[i].GetKey()
		if !caseSensitive {
			key = strings.ToLower(key)
		}
		if _, contains := hashTable[key]; !contains {
			hashTable[key] = envvarSet[i]
		}
	}
}
