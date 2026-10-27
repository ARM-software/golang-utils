<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Detect Go modules

`detect-go-modules` is a standalone composite action for discovering every
`go.mod` in a repository. If the repository has a root `go.work`, every module
must be listed there; otherwise discovery uses the repository layout. Missing
workspace entries fail the action instead of silently omitting packages.

| Output | Description |
| --- | --- |
| `modules` | Newline-delimited module directories relative to the repository root. |
| `modules-json` | JSON array for `fromJSON(...)` matrices or reusable workflow inputs. |

Check out the calling repository and set up Go before using the action:

```yaml
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- id: modules
  uses: ARM-software/golang-utils/actions/detect-go-modules@<full-commit-sha>
- run: echo '${{ steps.modules.outputs.modules-json }}'
```

Replace `<full-commit-sha>` with a commit from this repository. Inside this
repository use `uses: $/actions/detect-go-modules` so the action is taken from
the same revision as the running workflow.
