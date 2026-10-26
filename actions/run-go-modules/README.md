<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Run a command in Go modules

`run-go-modules` performs one operation in every selected module. It can run
independently (automatically discovering modules) or consume the `modules`
output of `detect-go-modules`.

| Input | Required | Description |
| --- | --- | --- |
| `task` | Yes | `build`, `download`, `analyze`, `generate`, or `lint`. |
| `build-tags` | No | Build tags for build, analysis, generation, and lint; defaults to `integration`. |
| `modules` | No | Newline-delimited module directories. Omit to discover all modules. |
| `results-file` | No | Repository-relative TSV containing each module's pass/fail/skipped status. |
| `skip-results-file` | No | Skip modules whose earlier task did not pass. |

For example, generate in every module of another repository:

```yaml
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- uses: ARM-software/golang-utils/actions/run-go-modules@<full-commit-sha>
  with:
    task: generate
    build-tags: integration
```

For `lint`, install `golangci-lint` first with
[`golangci-lint-action`](https://github.com/golangci/golangci-lint-action)
using `install-only: true`. `build`, `analyze`, and `generate`
pass `-tags integration` by default, while lint passes
`--build-tags integration`. `download` uses `GOWORK=off` to
exercise each module's own metadata. The action checks all selected modules
and fails if one fails; the optional TSV allows follow-on steps to report
individual outcomes. For calls from this repository use
`uses: $/actions/run-go-modules`.
