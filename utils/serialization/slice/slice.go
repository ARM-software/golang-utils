package slice

import (
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/serialization/maps"
)

// ToSlice converts a struct to a list of key values.
func ToSlice[T any](s *T) (list []string, err error) {
	r, err := maps.ToMap(s)
	if err != nil {
		return
	}
	list = collection.ConvertMapToSlice(r)
	return
}

// FromSlice converts a slice of key,values into a struct i
func FromSlice[T any](s []string, o *T) (err error) {
	m, err := collection.ConvertSliceToMap(s)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrMarshalling, err, "could not convert slice to map so it can be converted into `%T`", o)
		return
	}
	err = maps.FromMap[T](m, o)
	return
}

// FromArgs converts a list of args into a struct o
func FromArgs[T any](o *T, args ...string) error {
	return FromSlice(args, o)
}
