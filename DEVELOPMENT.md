# Development and Testing
## Local Development
 
Since this is Go library there is no application to run.
 
### Running tests

To run all tests:
```bash
go test ./...
```

### Mocks

We use the builtin mock library and use [mockgen](https://github.com/golang/mock#running-mockgen) to generate the mock implementation of Interfaces for testing.

To generate mocks:
```bash
go generate ./...
```

We make use of `go:generate` comments to instruct which `mockgen` commands should be run.
 
### Coding style
 
`goimports` is enforced via a CI step.
It is up to the individual developer to ensure their change complies with this.

A [pre-commit](https://pre-commit.com/) hook is provided to enforce this on `git commit`.
To enable this:
```bash
pre-commit install

# Test that it works
pre-commit run --all-files
```
 
Some additional static analysis tools are run in a CI step.
This uses [golangci-lint](https://golangci-lint.run/). To run this locally:
```bash
# Must be in a directory with a go.mod file
cd <directory_with_go_module>
golangci-lint run ./...
``` 

## Code Climate

Code Climate is integrated with our GitHub flow. Failing the configured rules will yield a pull request not mergeable.

If you prefer to view the Code Climate report on your machine, prior to sending a pull request, you can use the [cli provided by Code Climate](https://docs.codeclimate.com/docs/command-line-interface).

Plugins for various tools are also available:
  - [Atom](https://docs.codeclimate.com/docs/code-climate-atom-package)
  - [PyCharm](https://plugins.jetbrains.com/plugin/13306-code-cleaner-with-code-climate-cli)
  - [Vim](https://docs.codeclimate.com/docs/vim-plugin)
