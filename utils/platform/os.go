/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package platform

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

var (
	errNotSupportedByWindows = errors.New("not supported by windows")
	// https://learn.microsoft.com/en-us/previous-versions/troubleshoot/winautomation/product-documentation/best-practices/variables/percentage-character-usage-in-notations
	// https://ss64.com/nt/syntax-replace.html
	windowsVariableExpansionRegexStr = `%(?P<variable>[^:=]*)(:(?P<StrToFind>.*)=(?P<NewString>.*))?%`
	// UnixVariableNameRegexString defines the schema for variable names on Unix.
	// See https://www.gnu.org/software/bash/manual/bash.html#index-name and https://mywiki.wooledge.org/BashFAQ/006
	UnixVariableNameRegexString = "^[a-zA-Z_][a-zA-Z_0-9]*$"
	// WindowsVariableNameRegexString defines the schema for variable names on Windows.
	// See https://ss64.com/nt/syntax-variables.html
	WindowsVariableNameRegexString = "^[A-Za-z#$'()*+,.?@\\[\\]_`{}~][A-Za-z0-9#$'()*+,.?@\\[\\]_`{}~\\s]*$"
	errVariableNameInvalid         = validation.NewError("validation_is_variable_name", "must be a valid variable name")
	// IsWindowsVariableName defines a validation rule for variable names on Windows for use with github.com/go-ozzo/ozzo-validation
	IsWindowsVariableName = validation.NewStringRuleWithError(isWindowsVarName, errVariableNameInvalid)
	// IsUnixVariableName defines a validation rule for variable names on Unix for use with github.com/go-ozzo/ozzo-validation
	IsUnixVariableName = validation.NewStringRuleWithError(isUnixVarName, errVariableNameInvalid)
	// IsVariableName defines a validation rule for variable names for use with github.com/go-ozzo/ozzo-validation
	IsVariableName = validation.NewStringRuleWithError(isVarName, errVariableNameInvalid)
)

const (
	// PathListSeparator corresponds to the platform separator used to separate element in PATH environment variable
	// i.e. this character is used to separate filenames in a sequence of files given as a path list.
	// On UNIX systems, this character is ':'; on Microsoft Windows systems it is ';'.
	PathListSeparator rune = os.PathListSeparator
	// PathSeparator corresponds to the platform path separator as in the system-dependent default name-separator character.
	// On UNIX systems the value of this field is '/'; on Microsoft Windows systems it is '\\'.
	PathSeparator rune = os.PathSeparator
)

func isWindowsVarName(value string) bool {
	if validation.Required.Validate(value) != nil {
		return false
	}
	regex := regexp.MustCompile(WindowsVariableNameRegexString)
	return regex.MatchString(value)
}

func isUnixVarName(value string) bool {
	if validation.Required.Validate(value) != nil {
		return false
	}
	regex := regexp.MustCompile(UnixVariableNameRegexString)
	return regex.MatchString(value)
}

func isVarName(value string) bool {
	if IsWindows() {
		return isWindowsVarName(value)
	}
	return isUnixVarName(value)
}

// ConvertError converts a platform error into a commonerrors
func ConvertError(err error) error {
	switch {
	case err == nil:
		return err
	case commonerrors.Any(err, commonerrors.ErrTimeout, commonerrors.ErrCancelled):
		return err
	case commonerrors.Any(err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported):
		return err
	case IsWindows() && commonerrors.Any(err, errNotSupportedByWindows):
		return commonerrors.WrapError(commonerrors.ErrUnsupported, err, "")
	case commonerrors.CorrespondTo(err, "not supported"):
		return commonerrors.WrapError(commonerrors.ErrUnsupported, err, "")
	default:
		return err
		// TODO extend with more platform specific errors
	}
}

// IsWindows checks whether we are running on Windows or not.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// LineSeparator returns the line separator.
func LineSeparator() string {
	if IsWindows() {
		return "\r\n"
	}
	return UnixLineSeparator()
}

// UnixLineSeparator returns the line separator on Unix platform.
func UnixLineSeparator() string {
	return "\n"
}

// Hostname returns the hostname.
func Hostname() (string, error) {
	return os.Hostname()
}

// UpTime returns system uptime.
func UpTime() (uptime time.Duration, err error) {
	_uptime, err := host.Uptime()
	if err != nil {
		return
	}
	if _uptime > math.MaxInt64 { // The upper limit for time as defined by the standard library https://cs.opensource.google/go/go/+/master:src/time/time.go;l=915 is the same as math.MaxInt64
		err = commonerrors.Newf(commonerrors.ErrOutOfRange, "could not convert uptime '%v' to duration as it exceeds the upper limit for time.Duration", _uptime)
		return
	}
	uptime = time.Duration(safecast.ToInt64(_uptime)) * time.Second
	return
}

// BootTime returns system uptime.
func BootTime() (bootime time.Time, err error) {
	_bootime, err := host.BootTime()
	if err != nil {
		return
	}
	if _bootime > math.MaxInt64 { // The upper limit for time as defined by the standard library https://cs.opensource.google/go/go/+/master:src/time/time.go;l=915 is the same as math.MaxInt64
		err = commonerrors.Newf(commonerrors.ErrOutOfRange, "could not convert uptime '%v' to duration as it exceeds the upper limit for time.Duration", _bootime)
		return
	}
	bootime = time.Unix(safecast.ToInt64(_bootime), 0)
	return

}

// NodeName returns the system node name (equivalent to uname -n).
func NodeName() (nodename string, err error) {
	info, err := host.Info()
	if err != nil {
		return
	}
	nodename = fmt.Sprintf("%v (%v)", info.Hostname, info.HostID)
	return
}

// PlatformInformation returns the platform information (equivalent to uname -s).
func PlatformInformation() (information string, err error) {
	platform, family, version, err := host.PlatformInformation()
	if err != nil {
		return
	}
	information = fmt.Sprintf("%v (%v/%v)", platform, family, version)
	return
}

// SystemInformation returns the system information (equivalent to uname -a)
func SystemInformation() (information string, err error) {
	hostname, err := Hostname()
	if err != nil {
		return
	}
	nodename, err := NodeName()
	if err != nil {
		return
	}
	platform, err := PlatformInformation()
	if err != nil {
		return
	}
	uptime, err := UpTime()
	if err != nil {
		return
	}
	bootime, err := BootTime()
	if err != nil {
		return
	}
	information = fmt.Sprintf("Host: %v, Node: %v, Platform: %v, Up time: %v, Boot time: %v", hostname, nodename, platform, uptime, bootime)
	return
}

func Uname() (string, error) {
	return SystemInformation()
}

type RAM interface {
	// GetTotal returns total amount of RAM on this system
	GetTotal() uint64
	// GetAvailable returns RAM available for programs to allocate
	GetAvailable() uint64
	// GetUsed returns RAM used by programs
	GetUsed() uint64
	// GetUsedPercent returns Percentage of RAM used by programs
	GetUsedPercent() float64
	// GetFree returns kernel's notion of free memory
	GetFree() uint64
}

type VirtualMemory struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	Free        uint64
}

func (m *VirtualMemory) GetTotal() uint64        { return m.Total }
func (m *VirtualMemory) GetAvailable() uint64    { return m.Available }
func (m *VirtualMemory) GetUsed() uint64         { return m.Used }
func (m *VirtualMemory) GetUsedPercent() float64 { return m.UsedPercent }
func (m *VirtualMemory) GetFree() uint64         { return m.Free }

func GetRAM() (ram RAM, err error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	ram = &VirtualMemory{
		Total:       vm.Total,
		Available:   vm.Available,
		Used:        vm.Used,
		UsedPercent: vm.UsedPercent,
		Free:        vm.Free,
	}
	return
}

// SubstituteParameter performs parameter substitution on all platforms.
// - the first element is the parameter to substitute
// - if find and replace is also wanted, pass the pattern and the replacement as following arguments in that order.
func SubstituteParameter(parameter ...string) string {
	if IsWindows() {
		return SubstituteParameterWindows(parameter...)
	}
	return SubstituteParameterUnix(parameter...)
}

// SubstituteParameterUnix performs Unix parameter substitution:
// See https://tldp.org/LDP/abs/html/parameter-substitution.html
// - the first element is the parameter to substitute
// - if find and replace is also wanted, pass the pattern and the replacement as following arguments in that order.
func SubstituteParameterUnix(parameter ...string) string {
	if len(parameter) < 1 || !isUnixVarName(parameter[0]) { //nolint:gosec // golang order of evaluation means that this won't be out of range
		return "${}"
	}
	if len(parameter) < 3 || parameter[1] == "" {
		return fmt.Sprintf("${%v}", parameter[0])
	}
	return fmt.Sprintf("${%v//%v/%v}", parameter[0], parameter[1], parameter[2]) //nolint:gosec // by this point in the code parameter is verified to have at least 3 elements
}

// SubstituteParameterWindows performs Windows parameter substitution:
// See https://ss64.com/nt/syntax-replace.html
// - the first element is the parameter to substitute
// - if find and replace is also wanted, pass the pattern and the replacement as following arguments in that order.
func SubstituteParameterWindows(parameter ...string) string {
	if len(parameter) < 1 || !isWindowsVarName(parameter[0]) { //nolint:gosec // golang order of evaluation means that this won't be out of range
		return "%%"
	}
	if len(parameter) < 3 || parameter[1] == "" {
		return "%" + parameter[0] + "%"
	}
	return "%" + fmt.Sprintf("%v:%v=%v", parameter[0], parameter[1], parameter[2]) + "%" //nolint:gosec // by this point in the code parameter is verified to have at least 3 elements
}

// ExpandParameter expands a variable expressed in a string `s` with its value returned by the mapping function.
// If the mapping function returns a string with variables, it will expand them too if recursive is set to true.
func ExpandParameter(s string, mappingFunc func(string) (string, bool), recursive bool) string {
	if IsWindows() {
		return ExpandWindowsParameter(s, mappingFunc, recursive)
	}
	return ExpandUnixParameter(s, mappingFunc, recursive)
}

func newMappingFunc(recursive bool, mappingFunc func(string) (string, bool), expansionFunc func(s string, mappingFunc func(string) (string, bool)) string) func(string) (string, bool) {
	if recursive {
		return recursiveMapping(mappingFunc, expansionFunc)
	}
	return mappingFunc
}

func recursiveMapping(mappingFunc func(string) (string, bool), expansionFunc func(s string, mappingFunc func(string) (string, bool)) string) func(string) (string, bool) {
	newMappingFunc := func(entry string) (string, bool) {
		mappedEntry, found := mappingFunc(entry)
		if !found {
			return mappedEntry, found
		}
		newExpanded := expansionFunc(mappedEntry, mappingFunc)
		if mappedEntry == newExpanded {
			return newExpanded, true
		}
		return expansionFunc(newExpanded, mappingFunc), true
	}
	return newMappingFunc
}

// ExpandUnixParameter expands a ${param} or $param in `s` based on the mapping function
// See https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html
// os.Expand is used under the bonnet and so, only basic parameter substitution is performed.
// TODO if os.Expand is not good enough, consider using other libraries such as https://github.com/ganbarodigital/go_shellexpand or https://github.com/mvdan/sh
func ExpandUnixParameter(s string, mappingFunc func(string) (string, bool), recursive bool) string {
	mapping := newMappingFunc(recursive, mappingFunc, expandUnixParameter)
	return expandUnixParameter(s, mapping)
}

func expandUnixParameter(s string, mappingFunc func(string) (string, bool)) string {
	return os.Expand(s, func(variable string) string {
		mapped, _ := mappingFunc(variable)
		return mapped
	})
}

// ExpandWindowsParameter expands a %param% in `s` based on the mapping function
// See https://learn.microsoft.com/en-us/previous-versions/troubleshoot/winautomation/product-documentation/best-practices/variables/percentage-character-usage-in-notations
// https://devblogs.microsoft.com/oldnewthing/20060823-00/?p=29993
// https://github.com/golang/go/issues/24848
// WARNING: currently the function only works with one parameter substitution in `s`.
func ExpandWindowsParameter(s string, mappingFunc func(string) (string, bool), recursive bool) string {
	mapping := newMappingFunc(recursive, mappingFunc, expandWindowsParameter)
	return expandWindowsParameter(s, mapping)
}

func expandWindowsParameter(s string, mappingFunc func(string) (string, bool)) string {
	variableRegex := regexp.MustCompile(windowsVariableExpansionRegexStr)
	if !variableRegex.MatchString(s) {
		return s
	}
	allMatches := variableRegex.FindAllStringSubmatch(s, -1)
	expandedString := s
	for i := range allMatches {
		old, newStr := expandedVariableWithEdit(allMatches[i], mappingFunc)
		expandedString = strings.ReplaceAll(expandedString, old, newStr)
	}
	return expandedString
}

func expandedVariableWithoutEdit(match []string, mappingFunc func(string) (string, bool)) (string, string, bool) {
	if len(match) < 1 {
		return "", "", false
	}
	if len(match) < 2 {
		return match[0], "", false
	}
	variable := match[1]
	if len(strings.TrimSpace(variable)) == 0 {
		return match[0], match[0], false
	}
	expandedVariable, found := mappingFunc(variable)
	if found {
		return match[0], expandedVariable, true
	}
	return match[0], match[0], false
}
func expandedVariableWithEdit(match []string, mappingFunc func(string) (string, bool)) (string, string) {
	if len(match) != 5 {
		s, expandedVariable, _ := expandedVariableWithoutEdit(match, mappingFunc)
		return s, expandedVariable
	}
	strToFind := match[3]
	newString := match[4]
	s, expandedVariable, expanded := expandedVariableWithoutEdit(match, mappingFunc)
	if !expanded {
		return s, expandedVariable
	}
	return s, strings.ReplaceAll(expandedVariable, strToFind, newString)
}

// ExpandFromEnvironment expands a string containing variables with values from the environment.
// On unix, it is equivalent to os.ExpandEnv but differs on Windows due to the following issues:
// - https://learn.microsoft.com/en-gb/windows/win32/api/processenv/nf-processenv-expandenvironmentstringsa?redirectedfrom=MSDN
// - https://github.com/golang/go/issues/43763
// - https://github.com/golang/go/issues/24848
func ExpandFromEnvironment(s string, recursive bool) string {
	if IsWindows() {
		expanded := expandFromEnvironment(s)
		if recursive {
			newExpanded := expandFromEnvironment(expanded)
			if expanded == newExpanded {
				return expanded
			}
			return ExpandFromEnvironment(newExpanded, recursive)
		}
		return expanded
	}
	return ExpandUnixParameter(s, os.LookupEnv, recursive)
}
