# Go multimodules composite action

This action orchestrates the individually reusable
[`detect-go-modules`](../detect-go-modules/README.md),
[`run-go-modules`](../run-go-modules/README.md), and
[`test-go-modules`](../test-go-modules/README.md) composite actions. It checks
that no module is omitted and keeps generation, lint, tests, JUnit and coverage
reporting consistent across the repository. Each action may also be called
independently from another Go repository.

Tasks: `list`, `build`, `download`, `analyze`, `generate`, `lint`, `verify`.
`verify` runs generation, lint and `gotestsum` tests, uploads per-module JUnit
and coverage reports plus merged coverage, and writes the Actions summary.
Applicable tasks use the `integration` build tag by default; callers can
override it with `build-tags`.

Callers must check out the repository, set up Go, and install
`gotestsum` and `golangci-lint` when running `verify`. For example:

```yaml
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- run: |
    go install gotest.tools/gotestsum@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    echo "$(go env GOPATH)/bin" >> "$GITHUB_PATH"
- uses: ARM-software/golang-utils/actions/go-multimodules@<commit-sha>
  with:
    task: verify
    build-tags: integration
    artifact-prefix: '${{ github.job }}-'
```

Use a pinned commit SHA for the remote action reference. Within this repository,
use `uses: $/actions/go-multimodules`. `artifact-prefix` is
optional; its default keeps the artifact names `junit-xml` and `coverage`.

`modules-json` is available as an output for callers that need a matrix or
another workflow step. For separate-action examples, task modes, and how `$`
resolves across repositories, see
the [Go multimodules action guide](../../guides/action-go-multimodules.md).
