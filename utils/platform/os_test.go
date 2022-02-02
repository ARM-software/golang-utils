/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package platform

import (
	"fmt"
	"testing"

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
