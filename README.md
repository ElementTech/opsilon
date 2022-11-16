# Opsilon
[![Test](https://github.com/jatalocks/opsilon/actions/workflows/test.yml/badge.svg)](https://github.com/jatalocks/opsilon/actions/workflows/test.yml) [![golangci-lint](https://github.com/jatalocks/opsilon/actions/workflows/lint.yml/badge.svg)](https://github.com/jatalocks/opsilon/actions/workflows/lint.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/jatalocks/opsilon)](https://goreportcard.com/report/github.com/jatalocks/opsilon) [![Go Reference](https://pkg.go.dev/badge/github.com/jatalocks/opsilon.svg)](https://pkg.go.dev/github.com/jatalocks/opsilon) [![codecov](https://codecov.io/gh/jatalocks/opsilon/branch/main/graph/badge.svg?token=Y5K4SID71F)](https://codecov.io/gh/jatalocks/opsilon)

A customizable CLI for collaboratively running container-native workflows

For full usage, please refer to the: [Docs](/assets/doc.md).

<!--ts-->
   * [Quickstart](#quickstart)
   * [Demo](#demo)
   * [Contribution](#contribute)
     * [Development Features](###development-features)
     * [Project Layout](###project-layout)
     * [Makefile Targets](###makefile-targets)
<!--te-->



This project serves the purpose of giving developers, operations and other personal the ability to run custom workflows on their personal computer using a container environment, without them writing code and having to understand the meaning behind the script.
# Download
### Manually

Download the [latest release](https://github.com/jatalocks/opsilon/releases/latest) for your os: (this example uses version `v0.0.1`).
For Mac:
```bash
$ curl -L https://github.com/jatalocks/opsilon/releases/download/v0.0.1-alpha/opsilon_0.0.1-alpha_Darwin_x86_64.tar.gz \
 | tar -xz opsilon | chmod u+x opsilon
```
Test if the Opsilon CLI works: *(When it doesn't work, you may have downloaded the wrong file or your device/os isn't supported)*

```bash
$ ./opsilon version
```

Move the executable to a folder on your `$PATH`:

```bash
$ mv opsilon /usr/local/bin/opsilon # or /usr/bin/opsilon
```
## Usage
Make sure you have Docker installed on your computer.
1. Connect to the examples folder present in this repository
```sh
$ opsilon repo add --git -n examples -d examples -s examples/workflows -p https://github.com/jatalocks/opsilon.git -b main
```
2. List available workflows
```sh
$ opsilon list
```
3. Run a workflow!
```sh
$ opsilon run
```
# Demo

```sh
$> opsilon
A customizable CLI for collaboratively running container-native workflows

Usage:
  opsilon [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  list        List all workflows available in your repositories
  repo        Operate on workflow repositories
  run         Run an available workflow
  version     Displays opsilon version

Flags:
      --config string   config file (default is $HOME/.opsilon.yaml)
  -h, --help            help for opsilon

Use "opsilon [command] --help" for more information about a command.
```

# Contribution
I would always welcome an issue or a PR! Every contribution is welcome. Below is some information to help you get started.

### Development Features
- [goreleaser](https://goreleaser.com/) with `deb.` and `.rpm` package releasing
- [golangci-lint](https://golangci-lint.run/) for linting and formatting
- [Github Actions](.github/worflows) Stages (Lint, Test, Build, Release)
- [Gitlab CI](.gitlab-ci.yml) Configuration (Lint, Test, Build, Release)
- [cobra](https://cobra.dev/) CLI parser
- [Makefile](Makefile) - with various useful targets and documentation (see Makefile Targets)
- [Github Pages](_config.yml) using [jekyll-theme-minimal](https://github.com/pages-themes/minimal) (checkout [https://jatalocks.github.io/opsilon/](https://jatalocks.github.io/opsilon/))
- [pre-commit-hooks](https://pre-commit.com/) for formatting and validating code before committing

### Project Layout
* [assets/](https://pkg.go.dev/github.com/jatalocks/opsilon/assets) => docs
* [cmd/](https://pkg.go.dev/github.com/jatalocks/opsilon/cmd)  => commandline configurartions (flags, subcommands)
* [pkg/](https://pkg.go.dev/github.com/jatalocks/opsilon/pkg)  => the entrypoints of the CLI commands
* [internal/](https://pkg.go.dev/github.com/jatalocks/opsilon/pkg)  => packages that are the main core function of the project
- [`tools/`](tools/) => for automatically shipping all required dependencies when running `go get` (or `make bootstrap`) such as `golang-ci-lint` (see: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
)

### Makefile Targets
```sh
$> make
bootstrap                      install build deps
build                          build golang binary
clean                          clean up environment
cover                          display test coverage
docker-build                   dockerize golang application
fmt                            format go files
help                           list makefile targets
install                        install golang binary
lint                           lint go files
pre-commit                     run pre-commit hooks
run                            run the app
test                           display test coverage
```

### Thanks

This project was made possible by https://github.com/FalcoSuessgott/golang-cli-template