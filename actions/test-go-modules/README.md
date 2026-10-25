# Test Go modules and merge coverage

`test-go-modules` runs `gotestsum` separately in every Go module, emits one
JUnit XML and coverage profile per module, combines the coverage profiles,
uploads both report artifacts, and publishes a linked Actions step summary.
It can be used without the top-level `go-multimodules` action.

| Input | Required | Description |
| --- | --- | --- |
| `modules` | No | Newline-delimited module directories from `detect-go-modules`; omit to discover all. |
| `generation-results` | No | Repository-relative TSV from a preceding `run-go-modules` generate task. Failed generation skips that module's tests. |
| `lint-results` | No | Repository-relative TSV from a preceding lint task, for the summary. |
| `artifact-prefix` | No | Prefix for JUnit and coverage artifact names. |

The `combined-coverage` output contains the overall statement-coverage
percentage. A test failure causes the action to fail, but available JUnit and
coverage artifacts are still uploaded, and the summary lists failing tests.

To run this action from another repository, check out that repository, set up
Go, install `gotestsum` and make it available on `PATH`:

```yaml
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- run: |
    go install gotest.tools/gotestsum@latest
    echo "$(go env GOPATH)/bin" >> "$GITHUB_PATH"
- uses: ARM-software/golang-utils/actions/test-go-modules@<full-commit-sha>
  with:
    artifact-prefix: '${{ github.job }}-'
```

Replace `<full-commit-sha>` with a commit from this repository. Inside this
repository use `uses: $/actions/test-go-modules`. For a pipeline of separate
generation, lint, and test actions, see the
[Go multimodules action guide](../../guides/action-go-multimodules.md).
