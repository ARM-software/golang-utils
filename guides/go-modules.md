<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Go modules

This repository uses a single repository-wide `go.work` file at the repository root and a set of Go modules under `utils/`.

## Consumer usage

Depend on the specific submodule that provides the package you import.

Examples:

```bash
go get github.com/ARM-software/golang-utils/utils/http@latest
go get github.com/ARM-software/golang-utils/utils/filesystem@latest
go get github.com/ARM-software/golang-utils/utils/validation@latest
```

> **Compatibility note:** The `github.com/ARM-software/golang-utils/utils`
> module remains a lightweight umbrella for compatibility and tooling. New
> consumers should depend on the narrower submodule they actually import.
> Submodule tags start with the next [CDS] project release; `v1.170.7` and
> earlier published only the original umbrella module.
> If using dependabot, the upgrade and change should be transparent.

## Workspace structure

- `go.work` lives at the repository root.
- `utils/` is the umbrella module.
- Each top-level consumable package under `utils/` has its own `go.mod`.
- Shared helper packages that are reused across module boundaries also have their own `go.mod` when keeping them separate prevents test-only dependencies from leaking into production modules.

CI uses the reusable `sync-go-workspace` action to run `go work sync`, then the
Go multimodules action to discover modules from `go.work`. It checks that every
`go.mod` is listed.

## Internal dependencies

Internal module dependencies must follow this pattern:

1. Add a normal `require` on the repository release version, for example `github.com/ARM-software/golang-utils/utils/commonerrors v1.170.7` while preparing the first multi-module release.
2. Add a local `replace` pointing at the sibling checkout path, for example `replace github.com/ARM-software/golang-utils/utils/commonerrors => ../commonerrors`.
3. Keep `replace` directives suitable for local development with `GOWORK=off`.

This keeps local development working before publication. [CDS] 4.8.0+ is responsible
for preparing the versioned internal requirements and tagging all workspace
modules as part of the single project release. Go consumers ignore dependency
modules' local `replace` directives and resolve versioned requirements from
published [CDS] tags.

## Adding a new module

1. Create the package directory under `utils/`.
2. Add a `go.mod` whose module path matches the import path.
3. Add the module path to the repository `go.work`.
4. Add versioned `require` plus local `replace` directives for internal dependencies.
5. If the module contains `go:generate` directives, use `go tool` and declare
   only the needed tools in that module's `tool` block.
6. Update `.github/dependabot.yml` so the new `go.mod` directory is covered.
7. Ensure CI still discovers the module from `go.work`.

If the module root would otherwise contain no Go source file, add a small `doc.go` so Go tooling can still address the module cleanly.

When splitting an existing package into a module, follow the
[backward-compatible module refactoring guide](migrations/refactoring-go-modules.md)
to preserve consumer imports and dependencies.

## Running checks across all modules

From the repository root:

```bash
go work sync
go work edit -json |
  sed -n 's/^[[:space:]]*"DiskPath": "\([^"]*\)".*$/\1/p' |
  while IFS= read -r module; do
    (cd "$module" && go generate -tags integration ./... && golangci-lint run --build-tags integration ./... && go test -tags integration ./...)
  done
```

The composite action's `verify` task runs those checks with `gotestsum`, then
produces per-module JUnit XML and coverage profiles, plus combined reports,
under `artifacts/`. Local commands use Bash on Linux and macOS; on Windows, use
WSL with Go and `golangci-lint` installed.
To confirm that one module builds without workspace resolution, run
`GOWORK=off go test ./...` from its directory.

## CI behaviour

The CI workflows call the local [Go multimodules composite action](action-go-multimodules.md)
with `uses: $/actions/go-multimodules`. It:

1. Reads the module list from the repository `go.work`.
2. Runs `go generate` per module.
3. Runs `golangci-lint` per module.
4. Runs tests per module with `gotestsum` and writes per-module and combined JUnit XML reports.
5. Writes one coverage profile per module and merges them into a combined profile.
6. Uploads JUnit and coverage artifacts.
7. Publishes a GitHub Actions summary listing each module's coverage, failed
   test cases, and links to the JUnit and coverage artifacts.

See the [Go multimodules action guide](action-go-multimodules.md) for a full
example using the action from another repository, including the difference
between `$` and a repository-qualified `uses:` reference.

## [CDS] releases

[CDS] remains the only project release mechanism.

- All modules share the same repository release version.
- Submodules are not versioned or released independently.
- `cd-tag-and-release` remains the release entry point.
- In [CDS] 4.8.0+, `cd-tag-and-release` updates versioned internal `require`
  directives and includes them in the shared release commit before tagging
  and publication. The repository's `release.yml` requires this version.

[CDS]: https://github.com/ARMmbed/continuous-delivery-scripts
