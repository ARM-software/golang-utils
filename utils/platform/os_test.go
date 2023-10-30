/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package platform

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostname(t *testing.T) {
	hostname, err := Hostname()
	require.Nil(t, err)
	assert.NotZero(t, hostname)
}

func TestNodeName(t *testing.T) {
	nodename, err := NodeName()
	require.Nil(t, err)
	assert.NotZero(t, nodename)
}

func TestPlatformInformation(t *testing.T) {
	platform, err := PlatformInformation()
	require.Nil(t, err)
	assert.NotZero(t, platform)
}

func TestBootTime(t *testing.T) {
	boottime, err := BootTime()
	require.Nil(t, err)
	assert.NotZero(t, boottime)
}

func TestUptime(t *testing.T) {
	uptime, err := UpTime()
	require.Nil(t, err)
	assert.NotZero(t, uptime)
}

func TestSystemInformation(t *testing.T) {
	uname, err := Uname()
	require.Nil(t, err)
	assert.NotZero(t, uname)
	fmt.Println(uname)
}

func TestMemoryInformation(t *testing.T) {
	ram, err := GetRAM()
	require.Nil(t, err)
	assert.NotZero(t, ram)
	fmt.Println(ram)
}

func TestExpandParameter(t *testing.T) {
	complexVar := faker.Username()
	complexVar2 := faker.Username()
	random := faker.Sentence()
	mapping := map[string]string{complexVar: "a test", "a": "b", complexVar2: "another test", "var1": "last test"}
	mappingFunc := func(entry string) (string, bool) {
		if entry == "" {
			return "", true
		}
		mapped, found := mapping[entry]
		return mapped, found
	}
	tests := []struct {
		expression                string
		expandedExpressionUnix    string
		expandedExpressionWindows string
	}{
		{
			expression:                "",
			expandedExpressionUnix:    "",
			expandedExpressionWindows: "",
		},
		{
			expression:                fmt.Sprintf("  %v  ", SubstituteParameterUnix()),
			expandedExpressionUnix:    "    ",
			expandedExpressionWindows: "  ${}  ",
		},
		{
			expression:                fmt.Sprintf("  %v  ", SubstituteParameterUnix(random)),
			expandedExpressionUnix:    "    ",
			expandedExpressionWindows: fmt.Sprintf("  %v  ", SubstituteParameterUnix(random)),
		},
		{
			expression:                "   $   ",
			expandedExpressionUnix:    "   $   ",
			expandedExpressionWindows: "   $   ",
		},
		{
			expression:                fmt.Sprintf("  %v  ", SubstituteParameterWindows()),
			expandedExpressionUnix:    fmt.Sprintf("  %v  ", SubstituteParameterWindows()),
			expandedExpressionWindows: "  %%  ",
		},
		{
			expression:                "   %  %   ",
			expandedExpressionUnix:    "   %  %   ",
			expandedExpressionWindows: "   %  %   ",
		},
		{
			expression:                "   %  %   ",
			expandedExpressionUnix:    "   %  %   ",
			expandedExpressionWindows: "   %  %   ",
		},
		{
			expression:                "   %:=  %   ",
			expandedExpressionUnix:    "   %:=  %   ",
			expandedExpressionWindows: "   %:=  %   ",
		},
		{
			expression:                "   %  :=  %   ",
			expandedExpressionUnix:    "   %  :=  %   ",
			expandedExpressionWindows: "   %  :=  %   ",
		},
		{
			expression:                fmt.Sprintf("   %v   ", SubstituteParameterWindows(random)),
			expandedExpressionUnix:    fmt.Sprintf("   %v   ", SubstituteParameterWindows(random)),
			expandedExpressionWindows: fmt.Sprintf("   %v   ", SubstituteParameterWindows(random)),
		},
		{
			expression:                SubstituteParameterUnix(complexVar),
			expandedExpressionUnix:    "a test",
			expandedExpressionWindows: SubstituteParameterUnix(complexVar),
		},
		{
			expression:                fmt.Sprintf("  %v   ", SubstituteParameterUnix(complexVar)),
			expandedExpressionUnix:    "  a test   ",
			expandedExpressionWindows: fmt.Sprintf("  %v   ", SubstituteParameterUnix(complexVar)),
		},
		{
			expression:                SubstituteParameterWindows(SubstituteParameterUnix(complexVar2)),
			expandedExpressionUnix:    "%another test%",
			expandedExpressionWindows: SubstituteParameterWindows(SubstituteParameterUnix(complexVar2)),
		},
		{
			expression:                fmt.Sprintf("a1234556() ${%v}a1234556() .", complexVar2),
			expandedExpressionUnix:    "a1234556() another testa1234556() .",
			expandedExpressionWindows: fmt.Sprintf("a1234556() ${%v}a1234556() .", complexVar2),
		},
		{
			expression:                `%` + complexVar + `%`,
			expandedExpressionUnix:    `%` + complexVar + `%`,
			expandedExpressionWindows: "a test",
		},
		{
			expression:                `  %` + complexVar + `%  `,
			expandedExpressionUnix:    `  %` + complexVar + `%  `,
			expandedExpressionWindows: "  a test  ",
		},
		{
			expression:                `a1234556()${} %` + complexVar2 + `%  a1234556()${} .$`,
			expandedExpressionUnix:    `a1234556() %` + complexVar2 + `%  a1234556() .$`,
			expandedExpressionWindows: "a1234556()${} another test  a1234556()${} .$",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.expression, func(t *testing.T) {
			assert.Equal(t, test.expandedExpressionUnix, ExpandUnixParameter(test.expression, mappingFunc, false))
			assert.Equal(t, test.expandedExpressionWindows, ExpandWindowsParameter(test.expression, mappingFunc, false))
			if IsWindows() {
				assert.Equal(t, test.expandedExpressionWindows, ExpandParameter(test.expression, mappingFunc, false))
			} else {
				assert.Equal(t, test.expandedExpressionUnix, ExpandParameter(test.expression, mappingFunc, false))
			}
		})
	}
	assert.Equal(t, "a test", ExpandParameter(SubstituteParameter(complexVar), mappingFunc, true))
}

func TestRecursiveExpandParameter(t *testing.T) {
	complexVar := faker.Username()
	complexVar2 := faker.Name()
	random := faker.Sentence()
	windowsMapping := map[string]string{complexVar: "123456 %" + complexVar2 + "% .123-_", "a": "b", complexVar2: random, "var1": "last test"}
	linuxMapping := map[string]string{complexVar: "123456 ${" + complexVar2 + "} .123-_", "a": "b", complexVar2: random, "var1": "last test"}
	mappingFunc := func(windows bool) func(entry string) (string, bool) {
		return func(entry string) (string, bool) {
			if entry == "" {
				return "", true
			}
			var mapping map[string]string
			if windows {
				mapping = windowsMapping
			} else {
				mapping = linuxMapping
			}
			mapped, found := mapping[entry]
			return mapped, found

		}
	}
	tests := []struct {
		expression                string
		expandedExpressionUnix    string
		expandedExpressionWindows string
	}{
		{
			expression:                "",
			expandedExpressionUnix:    "",
			expandedExpressionWindows: "",
		},
		{
			expression:                "",
			expandedExpressionUnix:    "",
			expandedExpressionWindows: "",
		},
		{
			expression:                fmt.Sprintf("12345${%v} 123456", complexVar),
			expandedExpressionUnix:    fmt.Sprintf("12345123456 %v .123-_ 123456", random),
			expandedExpressionWindows: fmt.Sprintf("12345${%v} 123456", complexVar),
		},
		{
			expression:                `12345%` + complexVar + `% 123456`,
			expandedExpressionUnix:    `12345%` + complexVar + `% 123456`,
			expandedExpressionWindows: fmt.Sprintf("12345123456 %v .123-_ 123456", random),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.expression, func(t *testing.T) {
			assert.Equal(t, test.expandedExpressionUnix, ExpandUnixParameter(test.expression, mappingFunc(false), true))
			assert.Equal(t, test.expandedExpressionWindows, ExpandWindowsParameter(test.expression, mappingFunc(true), true))
			if IsWindows() {
				assert.Equal(t, test.expandedExpressionWindows, ExpandParameter(test.expression, mappingFunc(true), true))
			} else {
				assert.Equal(t, test.expandedExpressionUnix, ExpandParameter(test.expression, mappingFunc(false), true))
			}
		})
	}
}

func TestExpandFromEnvironment(t *testing.T) {
	if IsWindows() {
		t.Run("on windows", func(t *testing.T) {
			assert.NotEmpty(t, ExpandFromEnvironment(SubstituteParameterWindows("WINDIR"), false))
			assert.NotEqual(t, SubstituteParameterWindows("WINDIR"), ExpandFromEnvironment(SubstituteParameterWindows("WINDIR"), false))
			assert.Equal(t, SubstituteParameterWindows(), ExpandFromEnvironment(SubstituteParameterWindows(), false))
			assert.Equal(t, SubstituteParameterUnix(), ExpandFromEnvironment(SubstituteParameterUnix(), false))
		})
	} else {
		t.Run("on linux", func(t *testing.T) {
			assert.NotEmpty(t, ExpandFromEnvironment("$HOME", false))
			assert.NotEmpty(t, ExpandFromEnvironment(SubstituteParameterUnix("HOME"), false))
			assert.NotEqual(t, "$HOME", ExpandFromEnvironment("$HOME", false))
			assert.NotEqual(t, SubstituteParameterUnix("HOME"), ExpandFromEnvironment(SubstituteParameterUnix("HOME"), false))
			assert.Empty(t, ExpandFromEnvironment(SubstituteParameterUnix(), false))
		})
	}
	random := faker.Sentence()
	assert.Equal(t, SubstituteParameterWindows(random), ExpandFromEnvironment(SubstituteParameterWindows(random), false))
	assert.Equal(t, random, ExpandFromEnvironment(random, false))
	assert.Equal(t, SubstituteParameterWindows(), ExpandFromEnvironment(SubstituteParameterWindows(), false))
}

func TestSubstituteParameter(t *testing.T) {
	require.Equal(t, "${}", SubstituteParameterUnix())
	random := faker.Username()
	require.Equal(t, fmt.Sprintf("${%v}", random), SubstituteParameterUnix(random))
	require.Equal(t, fmt.Sprintf("${%v}", random), SubstituteParameterUnix(random, faker.Word()))
	require.Equal(t, fmt.Sprintf("${%v//pattern/replacement}", random), SubstituteParameterUnix(random, "pattern", "replacement"))
	require.Equal(t, "%%", SubstituteParameterWindows())
	require.Equal(t, "%"+random+"%", SubstituteParameterWindows(random))
	require.Equal(t, "%"+random+"%", SubstituteParameterWindows(random, faker.Word()))
	require.Equal(t, "%"+random+":pattern=replacement%", SubstituteParameterWindows(random, "pattern", "replacement"))
	require.NotEmpty(t, SubstituteParameter())
	require.NotEmpty(t, SubstituteParameter(random))
}

func TestExpandWindows(t *testing.T) {
	mapping := map[string]string{"var1": "first replacement", "var2": "second replacement", "var3": fmt.Sprintf("1.%v 2.%v", SubstituteParameterWindows("var1"), SubstituteParameterWindows("var2"))}

	mappingFunc := func(entry string) (string, bool) {
		if entry == "" {
			return "", true
		}
		mapped, found := mapping[entry]
		return mapped, found
	}

	assert.Equal(t, "second replacement", ExpandWindowsParameter(SubstituteParameterWindows("var2"), mappingFunc, true))
	assert.Equal(t, "second test", ExpandWindowsParameter(SubstituteParameterWindows("var2", "replacement", "test"), mappingFunc, true))
	// FIXME tweak the Expand function ExpandWindowsParameter to handle multiple parameter substitution in a string such as var3. Then uncomment the following test.
	// assert.Equal(t, "1.first replacement 2.second replacement", ExpandParameter(SubstituteParameterWindows("var3"), mappingFunc, true))
}

func TestValidateVariableName(t *testing.T) {
	require.NoError(t, IsVariableName.Validate(faker.Username()))
	require.NoError(t, IsVariableName.Validate(faker.Word()))
	require.Error(t, IsVariableName.Validate("9"+faker.UUIDDigit()))
	require.NoError(t, IsVariableName.Validate(faker.Word()+strings.ReplaceAll(faker.UUIDDigit(), "-", "_")))
	require.NoError(t, IsWindowsVariableName.Validate(faker.Username()))
	require.NoError(t, IsWindowsVariableName.Validate(faker.DomainName()))
	require.Error(t, IsWindowsVariableName.Validate("9"+faker.UUIDDigit()))
	require.NoError(t, IsWindowsVariableName.Validate(faker.Word()+strings.ReplaceAll(faker.UUIDDigit(), "-", "_")))
	require.NoError(t, IsUnixVariableName.Validate(faker.Username()))
	require.Error(t, IsUnixVariableName.Validate(faker.DomainName()))
	require.Error(t, IsWindowsVariableName.Validate("9"+faker.UUIDDigit()))
	require.NoError(t, IsUnixVariableName.Validate(faker.Word()+strings.ReplaceAll(faker.UUIDDigit(), "-", "_")))
}
