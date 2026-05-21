package easyjson

import (
	"slices"
	"time"

	"github.com/ARM-software/golang-utils/utils/collection"
)

//go:generate go tool easyjson -all $GOFILE

type TestingStruct struct {
	Int      int
	Int8     int8
	Int16    int16
	Int32    int32
	Int64    int64
	String   string
	Bool     bool
	Duration time.Duration
	SString  []string
	SInt     []int
	SInt8    []int8
	SInt16   []int16
	SInt32   []int32
	SInt64   []int64
	SFloat32 []float32
	SFloat64 []float64
	SBool    []bool
	Struct   AStruct
}

func (s *TestingStruct) Equals(other any) bool {
	if s == other {
		return true
	}

	if other == nil {
		return s == nil
	}

	if o, ok := other.(*TestingStruct); ok {
		return collection.All([]bool{
			s.Int == o.Int,
			s.Int8 == o.Int8,
			s.Int16 == o.Int16,
			s.Int32 == o.Int32,
			s.Int64 == o.Int64,
			s.String == o.String,
			s.Bool == o.Bool,
			s.Duration == o.Duration,
			slices.Equal(s.SString, o.SString),
			slices.Equal(s.SInt, o.SInt),
			slices.Equal(s.SInt8, o.SInt8),
			slices.Equal(s.SInt16, o.SInt16),
			slices.Equal(s.SInt32, o.SInt32),
			slices.Equal(s.SInt64, o.SInt64),
			slices.Equal(s.SFloat32, o.SFloat32),
			slices.Equal(s.SFloat64, o.SFloat64),
			slices.Equal(s.SBool, o.SBool),
			s.Struct.Equals(o.Struct),
		})
	}

	if o, ok := other.(TestingStruct); ok {
		return collection.All([]bool{
			s.Int == o.Int,
			s.Int8 == o.Int8,
			s.Int16 == o.Int16,
			s.Int32 == o.Int32,
			s.Int64 == o.Int64,
			s.String == o.String,
			s.Bool == o.Bool,
			slices.Equal(s.SString, o.SString),
			slices.Equal(s.SInt, o.SInt),
			slices.Equal(s.SInt8, o.SInt8),
			slices.Equal(s.SInt16, o.SInt16),
			slices.Equal(s.SInt32, o.SInt32),
			slices.Equal(s.SInt64, o.SInt64),
			slices.Equal(s.SFloat32, o.SFloat32),
			slices.Equal(s.SFloat64, o.SFloat64),
			slices.Equal(s.SBool, o.SBool),
			s.Struct.Equals(o.Struct),
		})
	}

	return false
}

type AStruct struct {
	Number        int64
	Height        int64
	AnotherStruct BStruct
}

func (s *AStruct) Equals(other any) bool {
	if s == other {
		return true
	}

	if other == nil {
		return s == nil
	}

	if o, ok := other.(*AStruct); ok {
		return s.Number == o.Number && s.Height == o.Height && s.AnotherStruct.Equals(&o.AnotherStruct)
	}

	if o, ok := other.(AStruct); ok {
		return s.Number == o.Number && s.Height == o.Height && s.AnotherStruct.Equals(&o.AnotherStruct)
	}

	return false
}

type BStruct struct {
	Image string
}

func (s *BStruct) Equals(other any) bool {
	if s == other {
		return true
	}

	if other == nil {
		return s == nil
	}

	if o, ok := other.(*BStruct); ok {
		return s.Image == o.Image
	}

	if o, ok := other.(BStruct); ok {
		return s.Image == o.Image
	}

	return false
}
