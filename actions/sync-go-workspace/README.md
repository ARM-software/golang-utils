<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Sync Go workspace

`sync-go-workspace` runs `go work sync` at the checked-out repository root. Use
it after setting up Go and before running checks or release commands that need
the workspace modules synchronized. The repository must have a root `go.work`.

In this repository:

```yaml
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- uses: $/actions/sync-go-workspace
```

From another repository, replace the last step with
`uses: ARM-software/golang-utils/actions/sync-go-workspace@<full-commit-sha>`.
