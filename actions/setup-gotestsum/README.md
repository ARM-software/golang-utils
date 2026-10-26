# Set up gotestsum

Installs [gotest.tools/gotestsum](https://github.com/gotestyourself/gotestsum)
using `go install` and adds the Go binary directory to `GITHUB_PATH`, so later
steps can run `gotestsum` without an extra setup command.

Set up Go first. The optional `version` input defaults to `v1.13.0`:

```yaml
- uses: actions/setup-go@v7
  with:
    go-version: '1.26.1'
- uses: ARM-software/golang-utils/actions/setup-gotestsum@<full-commit-sha>
  with:
    version: v1.13.0
- run: gotestsum --version
```

Replace `<full-commit-sha>` with a commit from this repository. From workflows
in this repository use `uses: $/actions/setup-gotestsum` to select the same
revision as the workflow.
