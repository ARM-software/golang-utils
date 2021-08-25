<!--
Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Development and Testing
## Local Development
 
Since this is a Go library, there is no application to run.
 
### Running tests

To run all tests:
```bash
go test ./...
```

### Mocks

We use the built-in mock library and use [mockgen](https://github.com/golang/mock#running-mockgen) to generate the mock implementation of Interfaces for testing.

To generate mocks:
```bash
go generate ./...
```

We make use of `go:generate` comments to instruct which `mockgen` commands should be run.
 
### Coding style
 
`goimports` is enforced via a CI step.
It is up to the individual developer to ensure their change complies with this.

### Static analysis and linting

Static analysis tools and linters are run as part of CI.
They come from [golangci-lint](https://golangci-lint.run/). To run this locally:
```bash
# Must be in a directory with a go.mod file
cd <directory_with_go_module>
golangci-lint run ./...
``` 

### Precommit

A [pre-commit](https://pre-commit.com/) hook configuration is provided to enforce some tasks on `git commit`.
Run `pre-commit install` to install pre-commit into your git hooks. pre-commit will now run on every commit.

If you want to manually run all pre-commit hooks on a repository without creating hooks, run `pre-commit run --all-files`. 

To run individual hooks use `pre-commit run <hook_id>`.

## Code Climate

Code Climate is integrated with our GitHub flow. Failing the configured rules will yield a pull request not mergeable.

If you prefer to view the Code Climate report on your machine, prior to sending a pull request, you can use the [cli provided by Code Climate](https://docs.codeclimate.com/docs/command-line-interface).

Plugins for various tools are also available:
  - [Atom](https://docs.codeclimate.com/docs/code-climate-atom-package)
  - [PyCharm](https://plugins.jetbrains.com/plugin/13306-code-cleaner-with-code-climate-cli)
  - [Vim](https://docs.codeclimate.com/docs/vim-plugin)

# Releasing

## Release Types

The CI supports three release flows:

- `development` for snapshot releases
- `release` for stable releases
- `beta` for pre-releases


|   Type      |   Purpose   | Version Number Format | GitHub Release | News Files Deleted |
|-------------|-------------|-----------------------|:--------------:|:------------------:|
| Release     | General Availability | `<minor>.<major>.<patch>`                            | Yes | Yes |
| Beta        | Integration Testing  | `<minor>.<major>.<patch>-beta.<commit number>`       | Yes | No  |
| Development | Development Testing  | `<minor>.<major>.<patch>-dev+<git hash>`             | No  | No  |

> :warning: releases can be made from any branches but
> it is recommended that they are only made from the `master` branch.

### Release workflow

1. Navigate to the [GitHub Actions](https://github.com/ARM-software/golang-utils/actions/workflows/release.yml) page.
2. Select the **Run Workflow** button and type which kind of release you would like to make (i.e. release, beta or development).

### Version Numbers

The version number will be automatically calculated, based on the news files.

# Detecting secrets

So that no secrets are committed back to the repository, a combination of two tools are run in CI:
- [GitLeaks]() : Scans the git history for usual secrets (e.g. AWS keys, etc.)
- [detect-secrets](https://github.com/Yelp/detect-secrets): Scans only the current state of the repository for anything which can look like secrets (strings with high entropy)

For the latter, False positive keys are stored in the [baseline](./.secrets.baseline) which `detect-secrets` checks against when it runs

## Baseline & False positives

To flag individual false positives add comment `# pragma: allowlist secret` to line with secret

To add all suspected secrets in the repository (excluding ones with an allow secret comment), run `detect-secrets scan --all-files  --exclude-files '.*go\.sum$' --exclude-files '.*\.html$' --exclude-files '.*\.properties$' --exclude-files 'ci.yml' > .secrets.baseline`

If on Windows: then change the encoding of the .secrets.baseline file to UTF-8 then convert all `\` to `/` in the .secrets.baseline file
