<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Replacing ozzo `validation.Required`

`github.com/go-ozzo/ozzo-validation/v4` changed `Required` semantics in v4.4.0.
Zero-valued structs now count as empty, which means code that previously treated
decoded struct values as present can start failing after the upgrade.

Upstream references:

- ozzo-validation v4.4.0 release notes:
  <https://github.com/go-ozzo/ozzo-validation/releases/tag/v4.4.0>
- ozzo-validation README, "Empty Values":
  <https://github.com/go-ozzo/ozzo-validation#empty-values>

## What changed

Before ozzo v4.4.0, `validation.Required` treated most zero-valued structs as
present. After v4.4.0, zero-valued structs are considered empty.

This affects patterns like:

```go
type Flags struct {
	Enabled bool
}

type Config struct {
	Flags Flags
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Flags, validation.Required),
	)
}
```

With newer ozzo versions, `Config{Flags: Flags{}}` now fails `validation.Required`.

## Replacement in golang-utils

Use `github.com/ARM-software/golang-utils/utils/validation.IsRequired` instead.

`IsRequired` is an alias of `IsRequiredLegacy` and preserves the older ozzo
behavior that many configuration-style validations relied on:

- nil values are rejected
- empty strings are rejected
- `time.Time{}` is rejected
- zero-valued non-time structs are still treated as present

## Migration

Replace this:

```go
import validation "github.com/go-ozzo/ozzo-validation/v4"

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Flags, validation.Required),
	)
}
```

With this:

```go
import (
	ozzovalidation "github.com/go-ozzo/ozzo-validation/v4"
	utilsvalidation "github.com/ARM-software/golang-utils/utils/validation"
)

func (c Config) Validate() error {
	return ozzovalidation.ValidateStruct(&c,
		ozzovalidation.Field(&c.Flags, utilsvalidation.IsRequired),
	)
}
```

If your file already imports golang-utils validation helpers, the change is often
just:

```go
validation.Field(&cfg.Flags, validationutils.IsRequired)
```

instead of:

```go
validation.Field(&cfg.Flags, validation.Required)
```

## When to keep ozzo `validation.Required`

Keep ozzo's built-in `validation.Required` when you explicitly want zero values
to be rejected, such as:

- `false` must be rejected
- `0` must be rejected
- zero-valued structs should be treated as missing

## Rule choice summary

- Use ozzo `validation.Required` for strict non-empty semantics.
- Use `utils/validation.Required` for state-style presence semantics where `0`
  and `false` should still count as present.
- Use `utils/validation.IsRequired` to preserve pre-v4.4 ozzo behavior for
  zero-valued structs.
