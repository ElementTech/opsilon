# gitform
A general purpose project template for golang CLI applications

<!--ts-->
   * [gitform](#gitform)
   * [Features](#features)
   * [Project Layout](#project-layout)
   * [How to use this template](#how-to-use-this-template)
   * [Demo Application](#demo-application)
   * [Makefile Targets](#makefile-targets)
   * [Contribute](#contribute)

<!-- Added by: morelly_t1, at: Tue 10 Aug 2021 08:54:24 AM CEST -->

<!--te-->

[![Test](https://github.com/amitai-devops/gitform/actions/workflows/test.yml/badge.svg)](https://github.com/amitai-devops/gitform/actions/workflows/test.yml) [![golangci-lint](https://github.com/amitai-devops/gitform/actions/workflows/lint.yml/badge.svg)](https://github.com/amitai-devops/gitform/actions/workflows/lint.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/amitai-devops/gitform)](https://goreportcard.com/report/github.com/amitai-devops/gitform) [![Go Reference](https://pkg.go.dev/badge/github.com/amitai-devops/gitform.svg)](https://pkg.go.dev/github.com/amitai-devops/gitform) [![codecov](https://codecov.io/gh/amitai-devops/gitform/branch/main/graph/badge.svg?token=Y5K4SID71F)](https://codecov.io/gh/amitai-devops/gitform)

This template serves as a starting point for golang commandline applications it is based on golang projects that I consider high quality and various other useful blog posts that helped me understanding golang better.

# Features
- [goreleaser](https://goreleaser.com/) with `deb.` and `.rpm` package releasing
- [golangci-lint](https://golangci-lint.run/) for linting and formatting
- [Github Actions](.github/worflows) Stages (Lint, Test, Build, Release)
- [Gitlab CI](.gitlab-ci.yml) Configuration (Lint, Test, Build, Release)
- [cobra](https://cobra.dev/) example setup including tests
- [Makefile](Makefile) - with various useful targets and documentation (see Makefile Targets)
- [Github Pages](_config.yml) using [jekyll-theme-minimal](https://github.com/pages-themes/minimal) (checkout [https://falcosuessgott.github.io/gitform/](https://falcosuessgott.github.io/gitform/))
- [pre-commit-hooks](https://pre-commit.com/) for formatting and validating code before committing

# Project Layout
* [assets/](https://pkg.go.dev/github.com/amitai-devops/gitform/assets) => docs, images, etc
* [cmd/](https://pkg.go.dev/github.com/amitai-devops/gitform/cmd)  => commandline configurartions (flags, subcommands)
* [pkg/](https://pkg.go.dev/github.com/amitai-devops/gitform/pkg)  => packages that are okay to import for other projects
* [internal/](https://pkg.go.dev/github.com/amitai-devops/gitform/pkg)  => packages that are only for project internal purposes
- [`tools/`](tools/) => for automatically shipping all required dependencies when running `go get` (or `make bootstrap`) such as `golang-ci-lint` (see: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
)

# How to use this template
```sh
bash <(curl -s https://raw.githubusercontent.com/amitai-devops/gitform/main/install.sh)
```

# Demo Application

```sh
$> gitform
golang-cli project template demo application

Usage:
  gitform [flags]
  gitform [command]

Available Commands:
  example     example subcommand which adds or multiplies two given integers
  help        Help about any command
  version     Displays d4sva binary version

Flags:
  -h, --help   help for gitform

Use "gitform [command] --help" for more information about a command.
```

```sh
$> gitform example 2 5 --add
7

$> gitform example 2 5 --multiply
10
```

# Makefile Targets
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

# Contribute
If you find issues in that setup or have some nice features / improvements, I would welcome an issue or a PR :)
