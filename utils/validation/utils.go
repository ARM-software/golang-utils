package validation

import (
	"reflect"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	utilreflection "github.com/ARM-software/golang-utils/utils/reflection"
)

type objectAccessor struct {
	names           []string
	value           func(string) (any, bool)
	present         func(string) bool
	presentNonEmpty func(string) bool
}

// typedSequence converts a validated input into a typed slice.
//
// It accepts either regular arrays/slices or function-backed iter.Seq-style
// values and attempts to cast every element to T, returning an error when an
// element cannot be converted.
func typedSequence[T any](value any) ([]T, error) {
	if seq, ok := sequenceToSlice(value); ok {
		return collection.MapWithError(seq, func(item any) (T, error) {
			cast, castOK := item.(T)
			if !castOK {
				var zero T
				return zero, commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported sequence item type: %T", item)
			}
			return cast, nil
		})
	}
	v, isNil := validation.Indirect(value)
	if isNil {
		return nil, nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return nil, errArrayOrSliceRequired
	}
	return collection.MapWithError(collection.Range(0, rv.Len(), nil), func(i int) (T, error) {
		item, ok := rv.Index(i).Interface().(T)
		if !ok {
			var zero T
			return zero, commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported item type at index [%d]: %T", i, rv.Index(i).Interface())
		}
		return item, nil
	})
}

// sequenceToSlice detects function-backed iter.Seq-style values and eagerly
// collects their yielded items into a slice.
//
// It returns false when value does not have the expected iterator function
// signature.
func sequenceToSlice(value any) ([]any, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return nil, false
	}
	rt := rv.Type()
	if rt.NumIn() != 1 || rt.NumOut() != 0 {
		return nil, false
	}
	yieldType := rt.In(0)
	if yieldType.Kind() != reflect.Func || yieldType.NumIn() != 1 || yieldType.NumOut() != 1 || yieldType.Out(0).Kind() != reflect.Bool {
		return nil, false
	}
	items := make([]any, 0)
	yield := reflect.MakeFunc(yieldType, func(args []reflect.Value) []reflect.Value {
		items = append(items, args[0].Interface())
		return []reflect.Value{reflect.ValueOf(true)}
	})
	rv.Call([]reflect.Value{yield})
	return items, true
}

// objectValue extracts a reflect.Value for map or struct inputs.
//
// It follows pointer/interface indirection using ozzo-validation utilities and
// returns an error when the resulting value is neither a map nor a struct.
func objectValue(value any) (rv reflect.Value, isNil bool, err error) {
	v, isNil := validation.Indirect(value)
	if isNil {
		return reflect.Value{}, true, nil
	}
	rv = reflect.ValueOf(v)
	if rv.Kind() != reflect.Map && rv.Kind() != reflect.Struct {
		return reflect.Value{}, false, errMapRequired
	}
	return rv, false, nil
}

// objectSequence2ToAccessor detects function-backed iter.Seq2-style values and
// collects string-keyed properties into an accessor.
func objectSequence2ToAccessor(value any) (*objectAccessor, bool, error) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return nil, false, nil
	}
	rt := rv.Type()
	if rt.NumIn() != 1 || rt.NumOut() != 0 {
		return nil, false, nil
	}
	yieldType := rt.In(0)
	if yieldType.Kind() != reflect.Func || yieldType.NumIn() != 2 || yieldType.NumOut() != 1 || yieldType.Out(0).Kind() != reflect.Bool {
		return nil, false, nil
	}
	items := make(map[string]any)
	keys := make([]string, 0)
	yield := reflect.MakeFunc(yieldType, func(args []reflect.Value) []reflect.Value {
		key, ok := args[0].Interface().(string)
		if !ok {
			panic(commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported object key type: %T", args[0].Interface()))
		}
		if _, found := items[key]; !found {
			keys = append(keys, key)
		}
		items[key] = args[1].Interface()
		return []reflect.Value{reflect.ValueOf(true)}
	})
	var callErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				switch err := recovered.(type) {
				case error:
					callErr = err
				default:
					panic(recovered)
				}
			}
		}()
		rv.Call([]reflect.Value{yield})
	}()
	if callErr != nil {
		return nil, true, callErr
	}
	return &objectAccessor{
		names: keys,
		value: func(key string) (any, bool) {
			value, found := items[key]
			return value, found
		},
		present: func(key string) bool {
			_, found := items[key]
			return found
		},
		presentNonEmpty: func(key string) bool {
			value, found := items[key]
			return found && !utilreflection.IsEmpty(value)
		},
	}, true, nil
}

// objectProperties extracts a uniform object accessor for map, struct, or
// string-keyed iter.Seq2 inputs.
func objectProperties(value any) (props *objectAccessor, isNil bool, err error) {
	if props, ok, err := objectSequence2ToAccessor(value); ok || err != nil {
		return props, false, err
	}
	rv, isNil, err := objectValue(value)
	if err != nil || isNil {
		return nil, isNil, err
	}
	keys := objectPropertyNames(rv)
	return &objectAccessor{
		names: keys,
		value: func(key string) (any, bool) {
			return objectPropertyValue(rv, key)
		},
		present: func(key string) bool {
			return hasObjectProperty(rv, key)
		},
		presentNonEmpty: func(key string) bool {
			value, found := objectPropertyValue(rv, key)
			return found && !utilreflection.IsEmpty(value)
		},
	}, false, nil
}

// hasObjectProperty reports whether rv contains a named property.
//
// For maps, this means a matching key exists. For structs, this means a field
// with that name exists and is not considered empty by the repository's
// reflection helpers.
func hasObjectProperty(rv reflect.Value, key string) bool {
	switch rv.Kind() {
	case reflect.Map:
		return rv.MapIndex(reflect.ValueOf(key)).IsValid()
	case reflect.Struct:
		fieldValue := rv.FieldByName(key)
		return fieldValue.IsValid() && !utilreflection.IsEmpty(fieldValue.Interface())
	default:
		return false
	}
}

// objectPropertyValue retrieves the value of a named property from a map or
// struct.
//
// The returned boolean reports whether the property exists.
func objectPropertyValue(rv reflect.Value, key string) (any, bool) {
	switch rv.Kind() {
	case reflect.Map:
		value := rv.MapIndex(reflect.ValueOf(key))
		if !value.IsValid() {
			return nil, false
		}
		return value.Interface(), true
	case reflect.Struct:
		fieldValue := rv.FieldByName(key)
		if !fieldValue.IsValid() {
			return nil, false
		}
		return fieldValue.Interface(), true
	default:
		return nil, false
	}
}

// objectPropertyNames returns the set of property names exposed by a map or
// struct.
//
// For maps, only string keys are returned. For structs, exported field names are
// returned as declared on the type.
func objectPropertyNames(rv reflect.Value) []string {
	switch rv.Kind() {
	case reflect.Map:
		keys := rv.MapKeys()
		return collection.Filter(collection.Map(keys, func(key reflect.Value) string {
			if key.Kind() != reflect.String {
				return ""
			}
			return key.String()
		}), func(key string) bool { return key != "" })
	case reflect.Struct:
		result := make([]string, 0, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			result = append(result, rv.Type().Field(i).Name)
		}
		return result
	default:
		return nil
	}
}

// countPresentObjectProperties counts how many of the supplied keys are present
// and non-empty on a map or struct value.
func countPresentObjectProperties(rv reflect.Value, keys []string) int {
	switch rv.Kind() {
	case reflect.Map:
		return collection.CountBy(keys, func(key string) bool {
			mapValue := rv.MapIndex(reflect.ValueOf(key))
			if !mapValue.IsValid() {
				return false
			}
			return !utilreflection.IsEmpty(mapValue.Interface())
		})
	case reflect.Struct:
		return collection.CountBy(keys, func(key string) bool {
			fieldValue := rv.FieldByName(key)
			if !fieldValue.IsValid() {
				return false
			}
			return !utilreflection.IsEmpty(fieldValue.Interface())
		})
	default:
		return 0
	}
}

// countPresentProperties counts how many of the supplied keys are present and
// non-empty on a normalised object accessor.
func countPresentProperties(props *objectAccessor, keys []string) int {
	if props == nil {
		return 0
	}
	return collection.CountBy(keys, props.presentNonEmpty)
}

// objectPropertyNamesFromAccessor returns the property names from props.
func objectPropertyNamesFromAccessor(props *objectAccessor) []string {
	if props == nil {
		return nil
	}
	return slices.Clone(props.names)
}
