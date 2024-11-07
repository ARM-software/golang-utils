package safecast

// This file is highly inspired from https://pkg.go.dev/golang.org/x/exp/constraints

// ISignedInteger is an alias for all signed integers: int, int8, int16, int32, and int64 types.
type ISignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// IUnsignedInteger is an alias for all unsigned integers: uint, uint8, uint16, uint32, and uint64 types.
type IUnsignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// IInteger is an alias for the all unsigned and signed integers
type IInteger interface {
	ISignedInteger | IUnsignedInteger
}

// IFloat is an alias for the float32 and float64 types.
type IFloat interface {
	~float32 | ~float64
}

// INumber is an alias for all integers and floats
type INumber interface {
	IInteger | IFloat
}

// IConvertable is an alias for everything that can be converted
type IConvertable interface {
	INumber
}
