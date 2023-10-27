/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package platform

import (
	"fmt"
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
	complexVar2 := faker.Name()
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
			expression:                "   ${}   ",
			expandedExpressionUnix:    "      ",
			expandedExpressionWindows: "   ${}   ",
		},
		{
			expression:                fmt.Sprintf("  ${%v}  ", random),
			expandedExpressionUnix:    "    ",
			expandedExpressionWindows: fmt.Sprintf("  ${%v}  ", random),
		},
		{
			expression:                "   $   ",
			expandedExpressionUnix:    "   $   ",
			expandedExpressionWindows: "   $   ",
		},
		{
			expression:                "   %%   ",
			expandedExpressionUnix:    "   %%   ",
			expandedExpressionWindows: "   %%   ",
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
			expression:                "   %" + random + "%   ",
			expandedExpressionUnix:    "   %" + random + "%   ",
			expandedExpressionWindows: "   %" + random + "%   ",
		},
		{
			expression:                fmt.Sprintf("${%v}", complexVar),
			expandedExpressionUnix:    "a test",
			expandedExpressionWindows: fmt.Sprintf("${%v}", complexVar),
		},
		{
			expression:                fmt.Sprintf("  ${%v}   ", complexVar),
			expandedExpressionUnix:    "  a test   ",
			expandedExpressionWindows: fmt.Sprintf("  ${%v}   ", complexVar),
		},
		{
			expression:                `%` + fmt.Sprintf("${%v}", complexVar2) + `%`,
			expandedExpressionUnix:    "%another test%",
			expandedExpressionWindows: `%` + fmt.Sprintf("${%v}", complexVar2) + `%`,
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
			assert.NotEmpty(t, ExpandFromEnvironment("%WINDIR%", false))
			assert.NotEqual(t, "%WINDIR%", ExpandFromEnvironment("%WINDIR%", false))
			assert.Equal(t, "%%", ExpandFromEnvironment("%%", false))
			assert.Equal(t, "${}", ExpandFromEnvironment("${}", false))
		})
	} else {
		t.Run("on linux", func(t *testing.T) {
			assert.NotEmpty(t, ExpandFromEnvironment("$HOME", false))
			assert.NotEmpty(t, ExpandFromEnvironment("${HOME}", false))
			assert.NotEqual(t, "$HOME", ExpandFromEnvironment("$HOME", false))
			assert.NotEqual(t, "${HOME}", ExpandFromEnvironment("${HOME}", false))
			assert.Empty(t, ExpandFromEnvironment("${}", false))
		})
	}
	random := faker.Sentence()
	assert.Equal(t, "%"+random+"%", ExpandFromEnvironment("%"+random+"%", false))
	assert.Equal(t, random, ExpandFromEnvironment(random, false))
	assert.Equal(t, "%%", ExpandFromEnvironment("%%", false))
}
