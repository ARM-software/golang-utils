<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Refactor Go modules without breaking imports

Use this guide when moving a package into its own module or replacing an old
import path with a preferred one. The goal is to keep existing consumers
working while allowing new code to depend on smaller modules and a single
implementation.

## Keep an import path when splitting a module

Adding `go.mod` inside a package directory changes which module provides that
package, even when its import path stays the same. For example, a new module at
`utils/codec/json` can retain the import path
`example.com/project/utils/codec/json`. The parent module no longer contains
that subtree at the new revision.

1. Declare the existing package import path in the new `go.mod`. Add the module
   to the repository's root `go.work` and ensure module discovery includes it.
2. Give each internal dependency a versioned `require` and a relative local
   `replace`, including dependencies between newly split modules. Test from
   each module with `GOWORK=off`: workspace resolution alone can hide a missing
   requirement or replacement. Go consumers ignore replacements in dependency
   modules, so the required versions must be published together.
3. Account for consumers that previously required only the parent module.
   Their import path may be unchanged, but Go must still find the new nested
   module in the build list. Where the parent serves as a compatibility
   umbrella, have it require the nested module at the shared release version;
   otherwise document the new module requirement for consumers.
4. Keep `.github/dependabot.yml` coverage, generator `tool` declarations, and
   generated files with the module that now owns each package. Check that
   [CDS](https://github.com/ARMmbed/continuous-delivery-scripts) prepares and
   tags every module in the same project release.

## Preserve an old path when changing imports

If the preferred import path differs from a published one, retain the old
package at its old path. Put the implementation in the preferred package and
make the old package a thin compatibility layer:

```go
package oldapi

import preferred "example.com/project/utils/newapi"

type Decoder = preferred.Decoder

func NewDecoder() *Decoder { return preferred.NewDecoder() }
```

Use type aliases (`type Old = preferred.Type`) for exported types, including
interfaces and option types, so values from either import path have the same
Go type identity. Forward exported functions, including generic helpers, to
the preferred implementation. Constants can refer to the preferred constants.
Remove duplicate implementations and generators from the compatibility package;
keep its import path and document the preferred one. Keep the dependency
direction from old to preferred to avoid import cycles.

Go has no true alias for an exported variable. If consumers can reassign one,
copying its value into a second variable does not preserve shared state. Review
such APIs individually rather than assuming that a wrapper preserves their
behaviour.

## Verify compatibility

- Compile and test both the preferred and old imports, including callers that
  use values from the two packages together. Compile-time assignments can
  confirm that exported types really are aliases.
- Run `go generate`, lint, and tests for each affected module with the project's
  build tags. Check modules independently with `GOWORK=off go test ./...` and
  `GOWORK=off go mod download`.
- Verify that every `go.mod` appears in `go.work`, that internal requirements
  point to the shared release version, and that the release publishes the
  modules consumers need. See the [Go modules guide](../go-modules.md) for the
  repository's workspace and CI commands.

For an import-path migration in this repository, see the
[serialisation packages guide](serialisation.md).
