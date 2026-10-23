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

The `github.com/ARM-software/golang-utils/utils` module remains in the repository as an umbrella module for compatibility and tooling, but new consumers should depend on the narrower submodule they actually import.
Submodule tags start with the next CDS project release; `v1.170.7` and earlier
published only the original umbrella module.

## Workspace structure

- `go.work` lives at the repository root.
- `utils/` is the umbrella module.
- Each top-level consumable package under `utils/` has its own `go.mod`.
- Shared helper packages that are reused across module boundaries also have their own `go.mod` when keeping them separate prevents test-only dependencies from leaking into production modules.

CI and local multi-module workflows discover modules from `go.work` with `go work edit -json`.

## Internal dependencies

Internal module dependencies must follow this pattern:

1. Add a normal `require` on the repository release version, for example `github.com/ARM-software/golang-utils/utils/commonerrors v1.170.7` while preparing the first multi-module release.
2. Add a local `replace` pointing at the sibling checkout path, for example `replace github.com/ARM-software/golang-utils/utils/commonerrors => ../commonerrors`.
3. Keep `replace` directives suitable for local development with `GOWORK=off`.

This keeps local development working before publication. The release workflow uses
`cd-determine-version` and `prepare_go_release.py` to update these requirements
to the shared CDS project version before `cd-tag-and-release` tags the modules.
Go consumers ignore dependency modules' local `replace` directives and resolve
the versioned requirements from the published CDS tags.

## Adding a new module

1. Create the package directory under `utils/`.
2. Add a `go.mod` whose module path matches the import path.
3. Add the module path to the repository `go.work`.
4. Add versioned `require` plus local `replace` directives for internal dependencies.
5. If the module contains `go:generate` directives, declare only the needed tools in that module.
6. Update `.github/dependabot.yml` so the new `go.mod` directory is covered.
7. Ensure CI still discovers the module from `go.work`.

If the module root would otherwise contain no Go source file, add a small `doc.go` so Go tooling can still address the module cleanly.

## Running checks across all modules

From the repository root:

```powershell
go work sync
$modules = @((go work edit -json | ConvertFrom-Json).Use | ForEach-Object { $_.DiskPath })
foreach ($module in $modules) {
  Push-Location $module
  $env:GOWORK = 'off'
  go generate ./...
  golangci-lint run ./...
  gotestsum -- -race -covermode=atomic -coverprofile coverage.out -tags integration -timeout 30m ./...
  Pop-Location
}
```

## CI behaviour

The CI workflow:

1. Reads the module list from the repository `go.work`.
2. Runs `go generate` per module.
3. Runs `golangci-lint` per module.
4. Runs tests per module with `gotestsum` and writes one JUnit XML file per module.
5. Writes one coverage profile per module and merges them into a combined profile.
6. Uploads JUnit and coverage artifacts.
7. Publishes a GitHub Actions summary listing the modules checked and the combined coverage result.

## CDS releases

CDS remains the only project release mechanism.

- All modules share the same repository release version.
- Submodules are not versioned or released independently.
- `cd-tag-and-release` remains the release entry point.
- The release workflow updates versioned internal `require` directives with
  `prepare_go_release.py` before invoking `cd-tag-and-release`; CDS remains the
  sole tagging and publication mechanism for the project.
