package nofast

import serialisation "github.com/ARM-software/golang-utils/utils/serialisation/yaml/yamltest/nofast"

// TestingStruct is a YAML test value without generated methods.
type TestingStruct = serialisation.TestingStruct

// AStruct is a nested YAML test value.
type AStruct = serialisation.AStruct

// BStruct is a nested YAML test value.
type BStruct = serialisation.BStruct
