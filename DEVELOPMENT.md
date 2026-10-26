<!--
Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Development and Testing
## Local Development
 
Since this is a Go library, there is no application to run.

The repository root `go.work` defines the workspace. The `utils/` directory contains the umbrella module plus the individual submodules.

Detailed multi-module guidance lives in the [Go modules guide](guides/go-modules.md).
 
### Running tests

To run all checks across every module from the repository root:

```bash
go work sync
go work edit -json |
  sed -n 's/^[[:space:]]*"DiskPath": "\([^"]*\)".*$/\1/p' |
  while IFS= read -r module; do
    (cd "$module" && go generate -tags integration ./... && golangci-lint run --build-tags integration ./... && go test -tags integration ./...)
  done
```

CI's `go-multimodules` action writes per-module JUnit XML and coverage to
`artifacts/`. To run checks for one module, run the Go commands from its
directory with `GOWORK=off`.

### Mocks

We use [go.uber.org/mock/mockgen](https://github.com/uber-go/mock) to generate
mock implementations of interfaces. Each module declares only the generators
it needs in its `go.mod` `tool` block; `go:generate` runs them via `go tool`.

To generate mocks:
```bash
for module in $(go work edit -json | sed -n 's/^[[:space:]]*"DiskPath": "\([^"]*\)".*$/\1/p'); do
  (cd "$module" && go generate -tags integration ./...)
done
```

We use `go:generate` directives for `mockgen`, `enumer`, `easyjson`, and
`ffjson` in the modules that need them.
 
### Coding style
 
`goimports` is enforced via a CI step.
It is up to the individual developer to ensure their change complies with this.

### Static analysis and linting

Static analysis tools and linters are run as part of CI.
They come from [golangci-lint](https://golangci-lint.run/). To run this locally:
```bash
for module in $(go work edit -json | sed -n 's/^[[:space:]]*"DiskPath": "\([^"]*\)".*$/\1/p'); do
  (cd "$module" && golangci-lint run --build-tags integration ./...)
done
``` 

### Precommit

A [pre-commit](https://pre-commit.com/) hook configuration is provided to enforce some tasks on `git commit`.
Run `pre-commit install` to install pre-commit into your git hooks. pre-commit will now run on every commit.

If you want to manually run all pre-commit hooks on a repository without creating hooks, run `pre-commit run --all-files`. 

To run individual hooks use `pre-commit run <hook_id>`.

# Migrations

This section covers what to do when a migration is needed. The goal of this
library is to preserve backward compatibility and avoid breaking changes, so
existing users should not have to migrate to keep their code working. See the
[migration guides](guides/migrations/index.md) for specific transitions.

# Dependency upgrades

For dependency upgrades, dependabot is relied upon and news files are auto-generated in order to document such change. Nonetheless, due to a change in [GitHub actions](https://github.blog/changelog/2021-02-19-github-actions-workflows-triggered-by-dependabot-prs-will-run-with-read-only-permissions), secrets are not available in the build triggered by the pull request unless they are [re-run manually](https://docs.github.com/en/code-security/supply-chain-security/keeping-your-dependencies-updated-automatically/automating-dependabot-with-github-actions#manually-re-running-a-workflow). So please re-run every dependabot PR CI jobs.

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
3. `cd-tag-and-release` regenerates `docs/` and writes the release-time TPIP output as part of the release flow.
4. [CDS](https://github.com/ARMmbed/continuous-delivery-scripts) remains the single release mechanism for the whole repository even though the repository now contains multiple modules.
5. Internal module `require` directives are versioned for [CDS](https://github.com/ARMmbed/continuous-delivery-scripts) release preparation, while local `replace` directives keep unreleased sibling modules consumable with `GOWORK=off`.

### Version Numbers

The version number will be automatically calculated, based on the news files.

### Adding a module

When adding a new module, also update:

1. `go.work`
2. `.github/dependabot.yml`
3. Any module-level `go:generate` tool declarations
4. Documentation in [guides/go-modules.md](guides/go-modules.md)

## Third-party IP checks

Pull requests now run `cd-check-licence-compliance` in CI so TPIP issues are caught before merge.

# Detecting secrets

CI calls [CDS](https://github.com/ARMmbed/continuous-delivery-scripts)'s
`cd-detect-secrets`, which checks tracked files with
[detect-secrets](https://github.com/Yelp/detect-secrets) against the project's
recorded baseline. Run `cd-detect-secrets` locally after installing CDS to
perform the same check.

False positives are recorded in the [baseline](./.secrets.baseline), which
`detect-secrets` checks when it runs.

## Baseline & False positives

To flag individual false positives add comment `# pragma: allowlist secret` to line with secret

To add all suspected secrets in the repository (excluding ones with an allow secret comment), run `detect-secrets scan --all-files  --exclude-files '.*\.html$' --exclude-files '.*\.properties$'  --exclude-files '^\.git[\\/]' --exclude-files ".*go\.sum$" --exclude-files '.*codegen.*' --exclude-files 'workflows/.*'  > .secrets.baseline`

If on Windows: then change the encoding of the .secrets.baseline file to UTF-8 then convert all `\` to `/` in the .secrets.baseline file
