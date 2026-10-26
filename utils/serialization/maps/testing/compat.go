package testing

import serialisation "github.com/ARM-software/golang-utils/utils/serialisation/maps/testing"

// TestEnumWithUnmarshal is a test enum with text unmarshalling.
type TestEnumWithUnmarshal = serialisation.TestEnumWithUnmarshal

// TestEnumWithoutUnmarshal is a test enum without text unmarshalling.
type TestEnumWithoutUnmarshal = serialisation.TestEnumWithoutUnmarshal

const (
	// TestEnumStringVer0 is the first text representation.
	TestEnumStringVer0 = serialisation.TestEnumStringVer0
	// TestEnumStringVer1 is the second text representation.
	TestEnumStringVer1 = serialisation.TestEnumStringVer1
	// TestEnumWithUnmarshal0 is the first enum value.
	TestEnumWithUnmarshal0 TestEnumWithUnmarshal = serialisation.TestEnumWithUnmarshal0
	// TestEnumWithUnmarshal1 is the second enum value.
	TestEnumWithUnmarshal1 TestEnumWithUnmarshal = serialisation.TestEnumWithUnmarshal1
	// TestEnumWithoutUnmarshal0 is the first enum value without text unmarshalling.
	TestEnumWithoutUnmarshal0 TestEnumWithoutUnmarshal = serialisation.TestEnumWithoutUnmarshal0
	// TestEnumWithoutUnmarshal1 is the second enum value without text unmarshalling.
	TestEnumWithoutUnmarshal1 TestEnumWithoutUnmarshal = serialisation.TestEnumWithoutUnmarshal1
)

// ValidationFunc checks a test enum value.
func ValidationFunc(value any) error { return serialisation.ValidationFunc(value) }
