# Go multimodules composite action

This action discovers every `go.mod` from `go.work` (or, if there is no
workspace, from the repository layout), checks that no module is omitted, and
runs the requested task in each module. It is a self-contained composite
action with inline Bash and standard Go tooling, so it can also be referenced
from another Go repository.

Tasks: `list`, `build`, `download`, `analyze`, `generate`, `lint`, `verify`.
`verify` runs generation, lint and `gotestsum` tests, uploads per-module JUnit
and coverage reports plus merged coverage, and writes the Actions summary.

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
- uses: ARM-software/golang-utils/.github/actions/go-multimodules@<commit-sha>
  with:
    task: verify
    artifact-prefix: '${{ github.job }}-'
```

Use a pinned commit SHA for the remote action reference. Within this repository,
use `uses: ./.github/actions/go-multimodules` after checkout. `artifact-prefix` is
optional; its default keeps the artifact names `junit-xml` and `coverage`.
