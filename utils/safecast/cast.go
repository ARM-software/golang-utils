package safecast

import (
	"math"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/ptr"
)

// CastFunc defines a function that converts a value of type F to type T,
// where both types satisfy [IConvertable].
type CastFunc[F, T IConvertable] func(F) T

// Cast converts an [IConvertable] value to another [IConvertable] type
// using the supplied conversion function.
func Cast[F, T IConvertable](i F, f CastFunc[F, T]) T {
	return f(i)
}

// RobustCast converts an [IConvertable] value to the target type T by
// dispatching to the appropriate safe helper for T.
//
// Conversions to bounded integer and float32 types use the corresponding
// helper, so out-of-range values are clamped to the nearest valid boundary.
// If types are not supported, an error is returned.
func RobustCast[F, T IConvertable](i F) (c T, err error) {
	var zero T

	switch any(zero).(type) {
	case int:
		c = any(Cast(i, ToInt[F])).(T)
	case uint:
		c = any(Cast(i, ToUint[F])).(T)
	case int8:
		c = any(Cast(i, ToInt8[F])).(T)
	case uint8:
		c = any(Cast(i, ToUint8[F])).(T)
	case int16:
		c = any(Cast(i, ToInt16[F])).(T)
	case uint16:
		c = any(Cast(i, ToUint16[F])).(T)
	case int32:
		c = any(Cast(i, ToInt32[F])).(T)
	case uint32:
		c = any(Cast(i, ToUint32[F])).(T)
	case int64:
		c = any(Cast(i, ToInt64[F])).(T)
	case uint64:
		c = any(Cast(i, ToUint64[F])).(T)
	case float32:
		c = any(Cast(i, ToFloat32[F])).(T)
	case float64:
		c = any(Cast(i, ToFloat64[F])).(T)
	default:
		err = commonerrors.New(commonerrors.ErrUnsupported, "target type for casting is not supported")
	}
	return
}

// CastRef converts a reference to an [IConvertable] value to a reference
// to another [IConvertable] type using the supplied conversion function.
// A nil input returns nil.
func CastRef[F, T IConvertable](i *F, f CastFunc[F, T]) *T {
	if i == nil {
		return nil
	}
	return ptr.To(Cast[F, T](ptr.From(i), f))
}

// RobustCastRef converts a reference to an [IConvertable] value to a reference
// to the target type T by dispatching to the appropriate safe helper for T.
// A nil input returns nil.
// If types are not supported, an error is returned.
func RobustCastRef[F, T IConvertable](i *F) (c *T, err error) {
	if i == nil {
		c = nil
		return
	}

	v, err := RobustCast[F, T](ptr.From(i))
	if err != nil {
		return
	}
	c = ptr.To(v)
	return
}

// ToInt attempts to convert an [IConvertable] value to an int.
// If the converted value falls outside the range of int,
// the nearest boundary value is returned instead.
func ToInt[C IConvertable](i C) int {
	if lessThanLowerBoundary(i, math.MinInt) {
		return math.MinInt
	}
	if greaterThanUpperBoundary(i, math.MaxInt) {
		return math.MaxInt
	}
	return int(i)
}

// ToIntRef attempts to convert a reference to an [IConvertable] value
// to a reference to an int. A nil input returns nil.
func ToIntRef[C IConvertable](i *C) *int {
	return CastRef(i, ToInt[C])
}

// ToUint attempts to convert an [IConvertable] value to a uint.
// If the converted value falls outside the range of uint,
// the nearest boundary value is returned instead.
func ToUint[C IConvertable](i C) uint {
	if lessThanLowerBoundary(i, uint(0)) {
		return 0
	}
	if greaterThanUpperBoundary(i, uint(math.MaxUint)) {
		return math.MaxUint
	}
	return uint(i)
}

// ToUintRef attempts to convert a reference to an [IConvertable] value
// to a reference to a uint. A nil input returns nil.
func ToUintRef[C IConvertable](i *C) *uint {
	return CastRef(i, ToUint[C])
}

// ToInt8 attempts to convert an [IConvertable] value to an int8.
// If the converted value falls outside the range of int8,
// the nearest boundary value is returned instead.
func ToInt8[C IConvertable](i C) int8 {
	if lessThanLowerBoundary(i, math.MinInt8) {
		return math.MinInt8
	}
	if greaterThanUpperBoundary(i, math.MaxInt8) {
		return math.MaxInt8
	}
	return int8(i)
}

// ToInt8Ref attempts to convert a reference to an [IConvertable] value
// to a reference to an int8. A nil input returns nil.
func ToInt8Ref[C IConvertable](i *C) *int8 {
	return CastRef(i, ToInt8[C])
}

// ToUint8 attempts to convert an [IConvertable] value to a uint8.
// If the converted value falls outside the range of uint8,
// the nearest boundary value is returned instead.
func ToUint8[C IConvertable](i C) uint8 {
	if lessThanLowerBoundary(i, 0) {
		return 0
	}
	if greaterThanUpperBoundary(i, math.MaxUint8) {
		return math.MaxUint8
	}
	return uint8(i)
}

// ToUint8Ref attempts to convert a reference to an [IConvertable] value
// to a reference to a uint8. A nil input returns nil.
func ToUint8Ref[C IConvertable](i *C) *uint8 {
	return CastRef(i, ToUint8[C])
}

// ToInt16 attempts to convert an [IConvertable] value to an int16.
// If the converted value falls outside the range of int16,
// the nearest boundary value is returned instead.
func ToInt16[C IConvertable](i C) int16 {
	if lessThanLowerBoundary(i, math.MinInt16) {
		return math.MinInt16
	}
	if greaterThanUpperBoundary(i, math.MaxInt16) {
		return math.MaxInt16
	}
	return int16(i)
}

// ToInt16Ref attempts to convert a reference to an [IConvertable] value
// to a reference to an int16. A nil input returns nil.
func ToInt16Ref[C IConvertable](i *C) *int16 {
	return CastRef(i, ToInt16[C])
}

// ToUint16 attempts to convert an [IConvertable] value to a uint16.
// If the converted value falls outside the range of uint16,
// the nearest boundary value is returned instead.
func ToUint16[C IConvertable](i C) uint16 {
	if lessThanLowerBoundary(i, 0) {
		return 0
	}
	if greaterThanUpperBoundary(i, math.MaxUint16) {
		return math.MaxUint16
	}
	return uint16(i)
}

// ToUint16Ref attempts to convert a reference to an [IConvertable] value
// to a reference to a uint16. A nil input returns nil.
func ToUint16Ref[C IConvertable](i *C) *uint16 {
	return CastRef(i, ToUint16[C])
}

// ToInt32 attempts to convert an [IConvertable] value to an int32.
// If the converted value falls outside the range of int32,
// the nearest boundary value is returned instead.
func ToInt32[C IConvertable](i C) int32 {
	if lessThanLowerBoundary(i, math.MinInt32) {
		return math.MinInt32
	}
	if greaterThanUpperBoundary(i, math.MaxInt32) {
		return math.MaxInt32
	}
	return int32(i)
}

// ToInt32Ref attempts to convert a reference to an [IConvertable] value
// to a reference to an int32. A nil input returns nil.
func ToInt32Ref[C IConvertable](i *C) *int32 {
	return CastRef(i, ToInt32[C])
}

// ToUint32 attempts to convert an [IConvertable] value to a uint32.
// If the converted value falls outside the range of uint32,
// the nearest boundary value is returned instead.
func ToUint32[C IConvertable](i C) uint32 {
	if lessThanLowerBoundary(i, 0) {
		return 0
	}
	if greaterThanUpperBoundary(i, math.MaxUint32) {
		return math.MaxUint32
	}
	return uint32(i)
}

// ToUint32Ref attempts to convert a reference to an [IConvertable] value
// to a reference to a uint32. A nil input returns nil.
func ToUint32Ref[C IConvertable](i *C) *uint32 {
	return CastRef(i, ToUint32[C])
}

// ToInt64 attempts to convert an [IConvertable] value to an int64.
// If the converted value falls outside the range of int64,
// the nearest boundary value is returned instead.
func ToInt64[C IConvertable](i C) int64 {
	if lessThanLowerBoundary(i, math.MinInt64) {
		return math.MinInt64
	}
	if greaterThanUpperBoundary(i, math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(i)
}

// ToInt64Ref attempts to convert a reference to an [IConvertable] value
// to a reference to an int64. A nil input returns nil.
func ToInt64Ref[C IConvertable](i *C) *int64 {
	return CastRef(i, ToInt64[C])
}

// ToUint64 attempts to convert an [IConvertable] value to a uint64.
// If the converted value falls outside the range of uint64,
// the nearest boundary value is returned instead.
func ToUint64[C IConvertable](i C) uint64 {
	if lessThanLowerBoundary(i, uint64(0)) {
		return 0
	}
	if greaterThanUpperBoundary(i, uint64(math.MaxUint64)) {
		return math.MaxUint64
	}
	return uint64(i)
}

// ToUint64Ref attempts to convert a reference to an [IConvertable] value
// to a reference to a uint64. A nil input returns nil.
func ToUint64Ref[C IConvertable](i *C) *uint64 {
	return CastRef(i, ToUint64[C])
}

// ToFloat64 attempts to convert an [IConvertable] value to a float64.
func ToFloat64[C IConvertable](i C) float64 {
	return float64(i)
}

// ToFloat64Ref attempts to convert a reference to an [IConvertable] value
// to a reference to a float64. A nil input returns nil.
func ToFloat64Ref[C IConvertable](i *C) *float64 {
	return CastRef(i, ToFloat64[C])
}

// ToFloat32 attempts to convert an [IConvertable] value to a float32.
// If the converted value falls outside the range of float32,
// the nearest boundary value is returned instead.
func ToFloat32[C IConvertable](i C) float32 {
	if lessThanLowerBoundary(i, -math.MaxFloat32) || math.IsInf(float64(i), -1) {
		return -math.MaxFloat32
	}
	if greaterThanUpperBoundary(i, math.MaxFloat32) || math.IsInf(float64(i), 1) {
		return math.MaxFloat32
	}
	return float32(i)
}

// ToFloat32Ref attempts to convert a reference to an [IConvertable] value
// to a reference to a float32. A nil input returns nil.
func ToFloat32Ref[C IConvertable](i *C) *float32 {
	return CastRef(i, ToFloat32[C])
}
