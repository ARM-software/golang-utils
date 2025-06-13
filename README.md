<!--
Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Golang utilities

[![Go Badge](https://img.shields.io/badge/go-v1.23-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI](https://github.com/ARM-software/golang-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/ARM-software/golang-utils/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ARM-software/golang-utils/utils.svg)](https://pkg.go.dev/github.com/ARM-software/golang-utils/utils)
[![Documentation](https://badgen.net/badge/Documentation/GitHub%20Pages/blue?icon=github)](https://arm-software.github.io/golang-utils)
[![Go Report Card](https://goreportcard.com/badge/github.com/ARM-software/golang-utils)](https://goreportcard.com/report/github.com/ARM-software/golang-utils)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6531/badge)](https://bestpractices.coreinfrastructure.org/projects/6531)
![Scorecard](https://img.shields.io/ossf-scorecard/github.com/ARM-software/golang-utils?label=openssf%20scorecard&style=flat)

## Overview
 
This repository contains various helpers and utilities for Golang in order to do actions not available in the standard library.
It was initially developed for numerous projects at Arm, including some running in production, in order to apply the DRY principle.

Few helpers may be themselves leveraging 3rd party libraries.
 
*Maintainers:* @ARM-software/golang-utils-admin 
 
## Using this library

To use this library, add the following line to your `go.mod`:
```
require (
    github.com/ARM-software/golang-utils/utils latest
    ...
)
```


## Releases

For release notes and a history of changes of all **production** releases, please see the following:

- [Changelog](CHANGELOG.md)

## Project Structure

The follow described the major aspects of the project structure:

- `docs/` - Code reference documentation.
- `utils/` - Go project source files.
- `changes/` - Collection of news files for unreleased changes.


## Getting Help

- For interface definition and usage documentation, please see [GitHub Pages](https://arm-software.github.io/golang-utils).
- For a list of known issues and possible workarounds, please see [Known Issues](KNOWN_ISSUES.md).
- To raise a defect or enhancement please use [GitHub Issues](https://github.com/ARM-software/golang-utils/issues) or [GitHub discussions](https://github.com/ARM-software/golang-utils/discussions).

## Contributing

- We are committed to fostering a welcoming community, please see our
  [Code of Conduct](CODE_OF_CONDUCT.md) for more information.
- For ways to contribute to the project, please see the [Contributions Guidelines](CONTRIBUTING.md)
- For a technical introduction into developing this package, please see the [Development Guide](DEVELOPMENT.md)




