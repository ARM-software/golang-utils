<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Use the Go multimodules action

The [`go-multimodules` composite action](https://github.com/ARM-software/golang-utils/tree/master/actions/go-multimodules)
orchestrates independently reusable actions for discovery, per-module commands,
and tests/coverage. They read `go.work` when present; otherwise they find
`go.mod` files directly. If a workspace omits a module, discovery fails rather
than silently skipping it. All actions use Bash and standard Go tooling. Some specific actions may rely on additional tools and dependencies though.

## Call it from this repository

This repository's workflows use:

```yaml
- uses: $/actions/go-multimodules
  with:
    task: build
```

`$/actions/...` resolves an action from the
**repository containing the workflow**, at the **same revision** as that
workflow. This keeps the workflow and action together when a reusable workflow
is called by another repository. It also lets a pull request exercise its own
changes to the action. The action lives at `actions/go-multimodules/action.yml`.

The top-level action composes `detect-go-modules`, `run-go-modules` (for
generation and lint), and `test-go-modules` (for JUnit and merged coverage).
Use it for the complete pipeline, or call any of its components on their own.

## Call it from another repository

Use a repository-qualified reference in the other repository's workflow. Pin
it to a full commit SHA or a version number available in `ARM-software/golang-utils`:

```yaml
name: Go modules
on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version: '1.26.1' # Choose the version required by your modules.
          cache-dependency-path: '**/go.sum'
      - uses: ARM-software/golang-utils/actions/sync-go-workspace@<full-commit-sha>
        if: ${{ hashFiles('go.work') != '' }}
      - uses: golangci/golangci-lint-action@v9
        with:
          version: latest
          install-only: true
      - uses: ARM-software/golang-utils/actions/setup-gotestsum@<full-commit-sha>
        with:
          version: v1.13.0
      - uses: ARM-software/golang-utils/actions/go-multimodules@<full-commit-sha>
        with:
          task: verify
          build-tags: integration
          artifact-prefix: '${{ github.job }}-'
```

Replace `<full-commit-sha>` with a real commit SHA or a version number from this repository. The
`$/actions/...` reference is for actions **in the workflow's repository**; in
the example above it would point at the calling repository, so the
repository-qualified reference is required instead. `artifact-prefix` is
optional; it helps prevent name collisions if several action invocations
upload reports in the same run.

## Choose an individual action

| Action | Standalone purpose | Useful inputs and outputs |
| --- | --- | --- |
| [`detect-go-modules`](https://github.com/ARM-software/golang-utils/tree/master/actions/detect-go-modules) | Verify that no module is omitted. | Outputs `modules` (newline-delimited) and `modules-json` for a matrix. |
| [`sync-go-workspace`](https://github.com/ARM-software/golang-utils/tree/master/actions/sync-go-workspace) | Run `go work sync` at the checked-out repository root. | No inputs; requires Go and `go.work`. |
| [`setup-gotestsum`](https://github.com/ARM-software/golang-utils/tree/master/actions/setup-gotestsum) | Install the `gotestsum` test runner and put it on `PATH`. | `version` defaults to `v1.13.0`. |
| [`run-go-modules`](https://github.com/ARM-software/golang-utils/tree/master/actions/run-go-modules) | Run one task across all modules. | Set `task` to `generate`, `lint`, `build`, `download`, or `analyze`; optionally supply `modules`, `results-file`, `skip-results-file`, and `build-tags`. |
| [`test-go-modules`](https://github.com/ARM-software/golang-utils/tree/master/actions/test-go-modules) | Test modules with `gotestsum`, merge coverage and publish reports. | Optionally supply `modules`, `generation-results`, `lint-results`, `artifact-prefix`, and `build-tags`; outputs `combined-coverage`. |
| [`go-multimodules`](https://github.com/ARM-software/golang-utils/tree/master/actions/go-multimodules) | Compose discovery, generation, lint and testing in one call. | Set `task` (default `verify`) and optionally `build-tags`; outputs `modules-json`. |

The individual actions can be used independently. If `modules` is omitted,
`run-go-modules` and `test-go-modules` call the discovery action themselves.
In this repository use `$/actions/<name>` to call each action; from another
repository use `ARM-software/golang-utils/actions/<name>@<full-commit-sha>`.

For example, to choose the stages separately in another repository, add the
following steps after checkout, Go setup, optional workspace sync, and the
`setup-gotestsum` and `golangci-lint-action` installation steps (as in the
complete example above):

```yaml
- id: modules
  uses: ARM-software/golang-utils/actions/detect-go-modules@<full-commit-sha>
- id: generate
  continue-on-error: true
  uses: ARM-software/golang-utils/actions/run-go-modules@<full-commit-sha>
  with:
    task: generate
    build-tags: integration
    modules: ${{ steps.modules.outputs.modules }}
    results-file: artifacts/generate.tsv
- id: lint
  if: ${{ always() }}
  continue-on-error: true
  uses: ARM-software/golang-utils/actions/run-go-modules@<full-commit-sha>
  with:
    task: lint
    build-tags: integration
    modules: ${{ steps.modules.outputs.modules }}
    results-file: artifacts/lint.tsv
    skip-results-file: artifacts/generate.tsv
- id: test
  if: ${{ always() }}
  uses: ARM-software/golang-utils/actions/test-go-modules@<full-commit-sha>
  with:
    build-tags: integration
    modules: ${{ steps.modules.outputs.modules }}
    generation-results: artifacts/generate.tsv
    lint-results: artifacts/lint.tsv
    artifact-prefix: '${{ github.job }}-'
```

The TSV results allow later stages to report generation and lint failures
without concealing test results. Use the same pinned SHA/version for every component
so their interfaces stay in sync. Build, analysis, generation, lint, and tests
include the `integration` build tag by default; `download` and discovery do not
compile tagged packages. Override `build-tags` if the calling project needs
additional tags.

## Tasks and reports

| Top-level `task` | Result |
| --- | --- |
| `list` | List the modules selected from `go.work` or `go.mod` files. |
| `download` | Run `go mod download` for each module with `GOWORK=off`. |
| `build` | Build every module. |
| `analyze` | Compile tests in every module without running test cases (for CodeQL). |
| `generate` | Run `go generate` in every module. |
| `lint` | Run `golangci-lint` in every module. |
| `verify` (default) | Generate, lint and test each module with `gotestsum`; collect coverage and reports. |

`verify` calls the granular generation, lint and test actions in order. The
test action uploads one JUnit XML file and one coverage profile per module,
plus a project-wide `combined.xml` JUnit report and `combined.out` coverage
profile. Its GitHub Actions summary shows every module's generate,
lint and test status, per-module coverage, failed test names, and links to the
uploaded reports. Artifact uploads and the summary also run when verification
fails, so available failure reports remain accessible.

Set up Go before using any task. For `verify` or `lint`, install
`golangci-lint`; for `verify`, also install `gotestsum`. GitHub-hosted runners
provide the Bash and standard Unix tools used by the action.
