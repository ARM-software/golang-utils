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

The legacy packages remain available so existing imports do
not need to change immediately. Their `doc.go` files mark them as deprecated though
and point to the newly created packages. In particular, the old
`serialization/xml` package delegates to `serialisation/xml`; legacy
compatibility tests cover the old XML, JSON, YAML and map import paths.

When migrating, change the import path and verify that any package-specific
types used together also come from the new import path.

Both module paths are listed in `go.work` and participate in one [CDS](https://github.com/ARMmbed/continuous-delivery-scripts)
release.
