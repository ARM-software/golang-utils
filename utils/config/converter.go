/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/value"
)

const (
	SensitiveFieldTagMask      = "mask"
	SensitiveFieldTagSensitive = "sensitive"
	SensitiveFieldTagSecret    = "secret"
	SensitiveFieldTagPassword  = "password"
	SensitiveFieldTagCensored  = "censored"
	SensitiveFieldTagRedact    = "redact"
	SensitiveFieldTagPII       = "pii"
	SensitiveFieldTagSens      = "sens"
)

// CommonSensitiveFieldTagNames lists common struct-tag keys used by Go
// libraries and applications to identify fields that should not be exposed
// verbatim in derived configuration views.
//
// These are tag keys checked via `reflect.StructTag.Lookup`, not field names.
// For example, this list is intended to match tags such as
// ``Password string `mask:"redact"``` or ``Email string `pii:"true"```, not
// field names like `Password`, `Secret`, or `Token` on their own.
//
// Recognising these tags helps reduce accidental disclosure of credentials,
// secrets, PII, and other confidential data in logs, environment-variable
// dumps, debug output, support bundles, and similar human-readable output.
//
// There is no standard Go struct tag for sensitive data. Some libraries use
// tags such as `sensitive`, `pii`, `sens`, `mask`, or `redact`, while major
// logging APIs such as `log/slog`, zap, and zerolog commonly address the
// problem through custom value or object marshaling instead.
//
// This mechanism can support privacy and data-protection controls such as data
// minimisation and pseudonymisation, but tagging or masking a field does not by
// itself establish compliance with GDPR or other privacy regulations.
//
// The tag choices are informed by patterns used in similar masking or redaction
// libraries and articles, including:
//   - https://github.com/ln80/struct-sensitive
//   - https://github.com/nirajbhattad/go-masker
//   - https://github.com/coopnorge/go-masker-lib/blob/main/catalog-info.yaml
//   - https://www.larcade.dev/articles/go-privacy-pii-redaction-masking
//   - https://github.com/ggwhite/go-masker
//   - https://github.com/showa-93/go-mask
//   - https://pkg.go.dev/github.com/anu1097/golang-masking-tool
//   - https://github.com/anu1097/golang-masking-tool
//   - https://dev.to/anu1097/introducing-golang-masking-tool-to-mask-away-your-sensitive-information-3387
var CommonSensitiveFieldTagNames = []string{
	SensitiveFieldTagMask,
	SensitiveFieldTagSensitive,
	SensitiveFieldTagSecret,
	SensitiveFieldTagPassword,
	SensitiveFieldTagCensored,
	SensitiveFieldTagRedact,
	SensitiveFieldTagPII,
	SensitiveFieldTagSens,
}

// NewValueTypeConverter wraps a reflection-style converter.
func NewValueTypeConverter(converter reflection.IValueConverter) value.IValueConverter {
	return reflection.NewValueTypeConverter(converter)
}

// NewValueMatchingConverter applies match to values of type T and replaces the
// value with the resulting boolean match outcome.
func NewValueMatchingConverter[T any](match collection.MatchingFunction[T]) value.IValueConverter {
	return value.ValueConverterFunc(func(ctx context.Context, value any) (any, error) {
		_ = ctx
		if match == nil {
			return value, nil
		}

		typedValue, ok := value.(T)
		if !ok {
			return value, nil
		}

		return match(typedValue), nil
	})
}

// SecretConverter converts a value to a masked string.
//
// It first uses [value.StringConverter] to normalise the input into a string,
// then preserves the first and last rune with `*****` in between.
var SecretConverter value.IValueConverter = value.ValueConverterFunc(func(ctx context.Context, input any) (any, error) {
	if reflection.IsEmpty(input) {
		return "", nil
	}
	converted, err := value.StringConverter.ConvertValue(ctx, input)
	if err != nil {
		return nil, err
	}
	text := fmt.Sprint(converted)
	runes := []rune(text)
	if len(runes) == 0 {
		return "", nil
	}
	first, _ := collection.First(runes)
	last, _ := collection.Last(runes)
	return fmt.Sprintf("%s*****%s", string(first), string(last)), nil
})

// NewFieldTagSecretConverter returns a converter that masks values whose source
// configuration field defines any of tagNames.
//
// It uses [reflection.StructPropertyHasTag] against the original configuration
// field metadata carried through the conversion context, and delegates the
// masking itself to [SecretConverter]. This is useful when generating
// environment-variable views or other derived representations of configuration
// that should avoid leaking secrets or sensitive personal data.
func NewFieldTagSecretConverter(tagNames ...string) value.IValueConverter {
	return value.ValueConverterFunc(func(ctx context.Context, input any) (any, error) {
		fieldContext, ok := currentConfigFieldContext(ctx)
		if !ok || len(tagNames) == 0 {
			return input, nil
		}
		if collection.AnyFunc(tagNames, func(tagName string) bool {
			return reflection.StructPropertyHasTag(fieldContext.owner, fieldContext.fieldName, tagName)
		}) {
			return SecretConverter.ConvertValue(ctx, input)
		}
		return input, nil
	})
}

// CommonFieldTagSecretConverter masks values whose source configuration field
// defines any tag listed in [CommonSensitiveFieldTagNames].
var CommonFieldTagSecretConverter value.IValueConverter = NewFieldTagSecretConverter(CommonSensitiveFieldTagNames...)

type configFieldContext struct {
	owner     reflect.Value
	fieldName string
}

type configFieldContextKey struct{}

func withConfigFieldContext(ctx context.Context, owner reflect.Value, fieldName string) context.Context {
	return context.WithValue(ctx, configFieldContextKey{}, configFieldContext{owner: owner, fieldName: fieldName})
}

func currentConfigFieldContext(ctx context.Context) (configFieldContext, bool) {
	fieldContext, ok := ctx.Value(configFieldContextKey{}).(configFieldContext)
	if !ok || !fieldContext.owner.IsValid() || fieldContext.fieldName == "" {
		return configFieldContext{}, false
	}
	return fieldContext, true
}
