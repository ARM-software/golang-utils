<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Test Go modules and merge coverage

`test-go-modules` runs `gotestsum` separately in every Go module, emits one
JUnit XML and coverage profile per module, combines the coverage profiles into
`combined.out`, uploads both report artifacts, and publishes JUnit test results
through `EnricoMi/publish-unit-test-result-action`.
It can be used without the top-level `go-multimodules` action.

| Input | Required | Description |
| --- | --- | --- |
| `build-tags` | No | Go test build tags; defaults to `integration`. |
| `modules` | No | Newline-delimited module directories from `detect-go-modules`; omit to discover all. |
| `generation-results` | No | Repository-relative TSV from a preceding `run-go-modules` generate task. Failed generation skips that module's tests. |
| `lint-results` | No | Repository-relative TSV from a preceding lint task, for the summary. |
| `artifact-prefix` | No | Prefix for JUnit and coverage artifact names. |

Tests run with `-tags integration` by default. The `combined-coverage` output
contains the overall statement-coverage percentage. A test failure causes the
action to fail, but available JUnit and coverage artifacts are still uploaded.
The JUnit artifact contains the original per-module reports, while the job
summary includes parsed test results plus a compact module and coverage table.
Grant the job `checks: write` if you want same-repository pushes and pull
requests to also receive a GitHub check run; fork pull requests still fall back
to the job summary only.

To run this action from another repository, check out that repository and
set up Go and [`gotestsum`](../setup-gotestsum/README.md):

```yaml
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- uses: ARM-software/golang-utils/actions/setup-gotestsum@<full-commit-sha>
- uses: ARM-software/golang-utils/actions/test-go-modules@<full-commit-sha>
  with:
    build-tags: integration
    artifact-prefix: '${{ github.job }}-'
```

Replace `<full-commit-sha>` with a commit from this repository. Inside this
repository use `uses: $/actions/test-go-modules`. For a pipeline of separate
generation, lint, and test actions, see the
[Go multimodules action guide](../../guides/action-go-multimodules.md).
