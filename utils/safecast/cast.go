package safecast

import "math"

// ToInt attempts to convert any [IConvertable] value to an int.
// If the conversion results in a value outside the range of an int,
// the closest boundary value will be returned.
func ToInt[C IConvertable](i C) int {
	if lessThanLowerBoundary(i, math.MinInt) {
		return math.MinInt
	}
	if greaterThanUpperBoundary(i, math.MaxInt) {
		return math.MaxInt
	}
	return int(i)
}

// ToUint attempts to convert any [IConvertable] value to an uint.
// If the conversion results in a value outside the range of an uint,
// the closest boundary value will be returned.
func ToUint[C IConvertable](i C) uint {
	if lessThanLowerBoundary(i, uint(0)) {
		return 0
	}
	if greaterThanUpperBoundary(i, uint(math.MaxUint)) {
		return math.MaxUint
	}
	return uint(i)
}

// ToInt8 attempts to convert any [IConvertable] value to an int8.
// If the conversion results in a value outside the range of an int8,
// the closest boundary value will be returned.
func ToInt8[C IConvertable](i C) int8 {
	if lessThanLowerBoundary(i, math.MinInt8) {
		return math.MinInt8
	}
	if greaterThanUpperBoundary(i, math.MaxInt8) {
		return math.MaxInt8
	}
	return int8(i)
}

// ToUint8 attempts to convert any [IConvertable] value to an uint8.
// If the conversion results in a value outside the range of an uint8,
// the closest boundary value will be returned.
func ToUint8[C IConvertable](i C) uint8 {
	if lessThanLowerBoundary(i, 0) {
		return 0
	}
	if greaterThanUpperBoundary(i, math.MaxUint8) {
		return math.MaxUint8
	}
	return uint8(i)
}

// ToInt16 attempts to convert any [IConvertable] value to an int16.
// If the conversion results in a value outside the range of an int16,
// the closest boundary value will be returned.
func ToInt16[C IConvertable](i C) int16 {
	if lessThanLowerBoundary(i, math.MinInt16) {
		return math.MinInt16
	}
	if greaterThanUpperBoundary(i, math.MaxInt16) {
		return math.MaxInt16
	}
	return int16(i)
}

// ToUint16 attempts to convert any [IConvertable] value to an uint16.
// If the conversion results in a value outside the range of an uint16,
// the closest boundary value will be returned.
func ToUint16[C IConvertable](i C) uint16 {
	if lessThanLowerBoundary(i, 0) {
		return 0
	}
	if greaterThanUpperBoundary(i, math.MaxUint16) {
		return math.MaxUint16
	}
	return uint16(i)
}

// ToInt32 attempts to convert any [IConvertable] value to an int32.
// If the conversion results in a value outside the range of an int32,
// the closest boundary value will be returned.
func ToInt32[C IConvertable](i C) int32 {
	if lessThanLowerBoundary(i, math.MinInt32) {
		return math.MinInt32
	}
	if greaterThanUpperBoundary(i, math.MaxInt32) {
		return math.MaxInt32
	}
	return int32(i)
}

// ToUint32 attempts to convert any [IConvertable] value to an uint32.
// If the conversion results in a value outside the range of an uint32,
// the closest boundary value will be returned.
func ToUint32[C IConvertable](i C) uint32 {
	if lessThanLowerBoundary(i, 0) {
		return 0
	}
	if greaterThanUpperBoundary(i, math.MaxUint32) {
		return math.MaxUint32
	}
	return uint32(i)
}

// ToInt64 attempts to convert any [IConvertable] value to an int64.
// If the conversion results in a value outside the range of an int64,
// the closest boundary value will be returned.
func ToInt64[C IConvertable](i C) int64 {
	if lessThanLowerBoundary(i, math.MinInt64) {
		return math.MinInt64
	}
	if greaterThanUpperBoundary(i, math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(i)
}

// ToUint64 attempts to convert any [IConvertable] value to an uint64.
// If the conversion results in a value outside the range of an uint64,
// the closest boundary value will be returned.
func ToUint64[C IConvertable](i C) uint64 {
	if lessThanLowerBoundary(i, uint64(0)) {
		return 0
	}
	if greaterThanUpperBoundary(i, uint64(math.MaxUint64)) {
		return math.MaxUint64
	}
	return uint64(i)
}

// ToFloat64 attempts to convert any [IConvertable] value to an float64.
func ToFloat64[C IConvertable](i C) float64 {
	return float64(i)
}
