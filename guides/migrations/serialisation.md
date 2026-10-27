<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Use serialisation packages

Use `github.com/ARM-software/golang-utils/utils/serialisation`
module for new applications. After the next project-wide
[CDS](https://github.com/ARMmbed/continuous-delivery-scripts) release, add it
to a Go project with:

```bash
go get github.com/ARM-software/golang-utils/utils/serialisation@latest
```

For example, import its XML helper as
`github.com/ARM-software/golang-utils/utils/serialisation/xml`. The same
spelling applies to the JSON, YAML, maps, slice and frontmatter subpackages:

| New import path suffix | Historical import path suffix |
| --- | --- |
| `/serialisation` | `/serialization` |
| `/serialisation/xml` | `/serialization/xml` |
| `/serialisation/json` | `/serialization/json` |
| `/serialisation/yaml` | `/serialization/yaml` |
| `/serialisation/maps` | `/serialization/maps` |
| `/serialisation/slice` | `/serialization/slice` |
| `/serialisation/frontmatter` | `/serialization/frontmatter` |
| `/serialisation/json/frontmatter` | `/serialization/json/frontmatter` |
| `/serialisation/yaml/frontmatter` | `/serialization/yaml/frontmatter` |
| `/serialisation/maps/testing` | `/serialization/maps/testing` |

The JSON and YAML frontmatter subpackages are separate Go modules. Depend on
the specific module you import, for example
`go get github.com/ARM-software/golang-utils/utils/serialisation/json/frontmatter@latest`.

If a consumer still requires a pre-split umbrella version such as
`github.com/ARM-software/golang-utils/utils v1.170.7`, upgrade that umbrella
requirement to the first split-aware release before adding one of the new
individual modules. Mixing a pre-split umbrella version with a split module
causes Go to report an `ambiguous import` error because both modules provide the
same package path.

The legacy packages remain available so existing imports do not need to change
immediately. Their `doc.go` files mark them as deprecated and point to the new
packages. All legacy helpers delegate to their `serialisation` counterparts;
exported types are aliases of the canonical types, including the frontmatter
format and parser types. Values can be used across the two import paths without
conversion.

When migrating, change the import path to its `serialisation` equivalent.

Both module paths are listed in `go.work` and participate in one [CDS](https://github.com/ARMmbed/continuous-delivery-scripts)
release.
