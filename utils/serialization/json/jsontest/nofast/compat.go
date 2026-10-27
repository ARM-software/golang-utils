/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package nofast

import serialisation "github.com/ARM-software/golang-utils/utils/serialisation/json/jsontest/nofast"

// TestingStruct is a JSON test value without generated methods.
type TestingStruct = serialisation.TestingStruct

// AStruct is a nested JSON test value.
type AStruct = serialisation.AStruct

// BStruct is a nested JSON test value.
type BStruct = serialisation.BStruct
