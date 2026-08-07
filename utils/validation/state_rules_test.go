package validation

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestStateRules(t *testing.T) {
	type zeroStruct struct {
		Enabled bool
	}

	nonZeroStruct := zeroStruct{Enabled: true}
	zeroTime := time.Time{}
	nonZeroTime := time.Now().UTC()

	t.Run("IsZero", func(t *testing.T) {
		require.NoError(t, validation.Validate(0, IsZero))
		require.NoError(t, validation.Validate(false, IsZero))
		require.NoError(t, validation.Validate(zeroStruct{}, IsZero))
		require.NoError(t, validation.Validate(zeroTime, IsZero))
		errortest.AssertErrorDescription(t, validation.Validate(1, IsZero), ErrZeroRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(nonZeroStruct, IsZero), ErrZeroRequired.Error())
	})

	t.Run("IsNotZero", func(t *testing.T) {
		require.NoError(t, validation.Validate(1, IsNotZero))
		require.NoError(t, validation.Validate(nonZeroStruct, IsNotZero))
		errortest.AssertErrorDescription(t, validation.Validate(0, IsNotZero), ErrNotZeroRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(zeroStruct{}, IsNotZero), ErrNotZeroRequired.Error())
	})

	t.Run("IsTrue and IsFalse", func(t *testing.T) {
		require.NoError(t, validation.Validate(true, IsTrue))
		require.NoError(t, validation.Validate(false, IsFalse))
		errortest.AssertErrorDescription(t, validation.Validate(false, IsTrue), ErrTrueRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(true, IsFalse), ErrFalseRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate("true", IsTrue), ErrTrueRequired.Error())
	})

	t.Run("IsEmpty and IsNotEmpty", func(t *testing.T) {
		require.NoError(t, validation.Validate("   ", IsEmpty))
		require.NoError(t, validation.Validate(false, IsEmpty))
		require.NoError(t, validation.Validate(zeroStruct{}, IsEmpty))
		errortest.AssertErrorDescription(t, validation.Validate("value", IsEmpty), ErrEmptyRequired.Error())

		require.NoError(t, validation.Validate("value", IsNotEmpty))
		require.NoError(t, validation.Validate(nonZeroStruct, IsNotEmpty))
		errortest.AssertErrorDescription(t, validation.Validate("", IsNotEmpty), ErrNotEmptyRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(zeroStruct{}, IsNotEmpty), ErrNotEmptyRequired.Error())
	})

	t.Run("IsNil and IsNotNil", func(t *testing.T) {
		var nilString *string
		nonNilString := "value"
		require.NoError(t, validation.Validate(nil, IsNil))
		require.NoError(t, validation.Validate(nilString, IsNil))
		errortest.AssertErrorDescription(t, validation.Validate(nonNilString, IsNil), ErrNilRequired.Error())

		require.NoError(t, validation.Validate(nonNilString, IsNotNil))
		require.NoError(t, validation.Validate(&nonNilString, IsNotNil))
		errortest.AssertErrorDescription(t, validation.Validate(nil, IsNotNil), ErrNotNilRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(nilString, IsNotNil), ErrNotNilRequired.Error())
	})

	t.Run("Required", func(t *testing.T) {
		var nilString *string
		var nilSlice []string
		var nilMap map[string]string
		emptyString := ""
		spaceString := "   "
		spaceStringPtr := &spaceString
		zeroTime := time.Time{}
		zeroTimePtr := &zeroTime
		zeroInt := 0
		falseBool := false
		emptySlice := []string{}
		emptyMap := map[string]string{}

		errortest.AssertErrorDescription(t, validation.Validate(nil, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(nilString, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(nilSlice, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(nilMap, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(emptyString, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(spaceString, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(spaceStringPtr, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(zeroTime, Required), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(zeroTimePtr, Required), validation.ErrRequired.Error())

		require.NoError(t, validation.Validate(zeroInt, Required))
		require.NoError(t, validation.Validate(falseBool, Required))
		require.NoError(t, validation.Validate(zeroStruct{}, Required))
		require.NoError(t, validation.Validate(emptySlice, Required))
		require.NoError(t, validation.Validate(emptyMap, Required))
	})

	t.Run("IsNotNilAndNotEmpty", func(t *testing.T) {
		var nilString *string
		emptyString := ""
		nonNilString := "value"
		require.NoError(t, validation.Validate(nonNilString, IsNotNilAndNotEmpty))
		require.NoError(t, validation.Validate(&nonNilString, IsNotNilAndNotEmpty))
		errortest.AssertErrorDescription(t, validation.Validate(nil, IsNotNilAndNotEmpty), ErrNotNilRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(nilString, IsNotNilAndNotEmpty), ErrNotNilRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(emptyString, IsNotNilAndNotEmpty), ErrNotEmptyRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(zeroStruct{}, IsNotNilAndNotEmpty), ErrNotEmptyRequired.Error())
	})

	t.Run("IsNilOrNotEmpty", func(t *testing.T) {
		var nilString *string
		nonNilString := "value"
		require.NoError(t, validation.Validate(nil, IsNilOrNotEmpty))
		require.NoError(t, validation.Validate(nilString, IsNilOrNotEmpty))
		require.NoError(t, validation.Validate(nonNilString, IsNilOrNotEmpty))
		errortest.AssertErrorDescription(t, validation.Validate("", IsNilOrNotEmpty), validation.ErrNilOrNotEmpty.Error())
		errortest.AssertErrorDescription(t, validation.Validate("   ", IsNilOrNotEmpty), validation.ErrNilOrNotEmpty.Error())
	})

	t.Run("IsRequiredLegacy preserves legacy ozzo struct semantics", func(t *testing.T) {
		errortest.AssertErrorDescription(t, validation.Validate(false, IsRequiredLegacy), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate("", IsRequiredLegacy), validation.ErrRequired.Error())
		errortest.AssertErrorDescription(t, validation.Validate(zeroTime, IsRequiredLegacy), validation.ErrRequired.Error())

		require.NoError(t, validation.Validate(zeroStruct{}, IsRequiredLegacy))
		require.NoError(t, validation.Validate(nonZeroStruct, IsRequiredLegacy))
		require.NoError(t, validation.Validate(nonZeroTime, IsRequiredLegacy))
	})

	t.Run("IsRequired aliases IsRequiredLegacy", func(t *testing.T) {
		require.NoError(t, validation.Validate(zeroStruct{}, IsRequired))
		errortest.AssertErrorDescription(t, validation.Validate(false, IsRequired), validation.ErrRequired.Error())
	})

	t.Run("ozzo required semantics on zero-valued structs stay pinned", func(t *testing.T) {
		type boolStruct struct {
			Enabled bool
			Ready   bool
		}
		type nestedStruct struct {
			Flags boolStruct
		}

		zeroBools := boolStruct{}
		zeroNested := nestedStruct{}

		// Pin ozzo's current behaviour so future upgrades make any semantic drift
		// explicit in tests before it changes repository validation expectations.
		require.NoError(t, validation.Validate(zeroBools, validation.Required))
		require.NoError(t, validation.Validate(zeroNested, validation.Required))
		errortest.AssertErrorDescription(t, validation.Validate(time.Time{}, validation.Required), validation.ErrRequired.Error())

		// Our legacy-compatible helper keeps zero-valued structs present even when
		// all their fields are zero values, such as booleans set to false.
		require.NoError(t, validation.Validate(zeroBools, IsRequiredLegacy))
		require.NoError(t, validation.Validate(zeroNested, IsRequiredLegacy))
		require.NoError(t, validation.Validate(zeroBools, IsRequired))
		require.NoError(t, validation.Validate(zeroNested, IsRequired))
	})

}
