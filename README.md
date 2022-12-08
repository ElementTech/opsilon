# Opsilon
[![Test](https://github.com/jatalocks/opsilon/actions/workflows/test.yml/badge.svg)](https://github.com/jatalocks/opsilon/actions/workflows/test.yml) [![golangci-lint](https://github.com/jatalocks/opsilon/actions/workflows/lint.yml/badge.svg)](https://github.com/jatalocks/opsilon/actions/workflows/lint.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/jatalocks/opsilon)](https://goreportcard.com/report/github.com/jatalocks/opsilon) [![Go Reference](https://pkg.go.dev/badge/github.com/jatalocks/opsilon.svg)](https://pkg.go.dev/github.com/jatalocks/opsilon) [![codecov](https://codecov.io/gh/jatalocks/opsilon/branch/main/graph/badge.svg?token=Y5K4SID71F)](https://codecov.io/gh/jatalocks/opsilon)

A customizable CLI for collaboratively running container-native workflows

![opsilon](https://user-images.githubusercontent.com/99724952/202414217-49f6a1f3-584d-4a6d-8fae-e92e888e1b86.svg)

For full usage, please refer to the: [Docs](/assets/doc.md).

<!--ts-->
- [Opsilon](#opsilon)
- [Download](#download)
    - [Quickstart](#quickstart)
    - [Helm](#helm)
  - [Usage](#usage)
    - [Extra Flags](#extra-flags)
- [Demo](#demo)
- [Contribution](#contribution)
    - [Development Features](#development-features)
    - [Project Layout](#project-layout)
    - [Makefile Targets](#makefile-targets)
    - [Thanks](#thanks)
<!--te-->



This project serves the purpose of giving developers, operations and other personal the ability to run custom workflows on their personal computer using a container environment, without them writing code and having to understand the meaning behind the script.
# Download
### Quickstart

Download the [latest release](https://github.com/jatalocks/opsilon/releases/latest) for your os: (this example uses version `v0.4.5`).
For Mac:
```bash
$ curl -L https://github.com/jatalocks/opsilon/releases/download/v0.4.5-alpha/opsilon_0.4.5-alpha_Darwin_x86_64.tar.gz \
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

### Helm

```bash
$ helm install https://github.com/jatalocks/opsilon/releases/download/opsilon-0.4.2-helm/opsilon-0.4.2-helm.tgz
```
## Usage
Make sure you have Docker installed on your computer (or connected to a kubernetes cluster `--kubernetes`).

 **EITHER**
1. Connect to the examples folder present in this repository
```sh
$ opsilon repo add --git -n examples -d examples -s examples/workflows -p https://github.com/jatalocks/opsilon.git -b main
# For private repositories, use https://myuser:github_token@github.com/myprivateorg/>myprivaterepo.git
```
2. List available workflows
```sh
$ opsilon list
```
3. Run a workflow!
```sh
$ opsilon run # --kubernetes (kubernetes instead of docker)
```
 **OR**
1. Start the web server
```sh
$ opsilon server -p 8080 # See extra flags below
```
2. List available API actions
```sh
$ Go to http://localhost:8080/api/v1/docs
```
 **OR**
1. Start the slack server
```sh
$ export SLACK_BOT_TOKEN=xoxb-123
$ export SLACK_APP_TOKEN=xapp-123
$ opsilon slack # See extra flags below
```
2. Install the app from the manifest in [Manifest](/assets/manifest.yaml)
3. Go to the Opsilon app in your slack
```sh
$ help

list repos - List Available Workflows
  Example: list
  Example: list myteam
  Example: list examples,myteam
run
```

### Extra Flags
`--kubernetes` - (kubernetes instead of docker)
___

**`server` or `slack` will usually come with:**

`--consul`  - Enable a consul server as a configuration endpoint. Allows for a distributed remote configuration of workflows and repositories instead of a local file.
   - `--consul_uri` = `localhost:8500` by default
   - `--consul_key` = `default` by default (which key to load configuration from)
  
`--database`  - Enable a mongodb database. Allows for logging and viewing workflow runs.
   - `--mongodb_uri` = `mongodb://localhost:27017` by default
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
  server      Runs an api server that functions the same as the CLI
  slack       Runs opsilon as a socket-mode slack bot
  version     Displays opsilon version

Flags:
      --config string        config file (default is $HOME/.opsilon.yaml)
      --consul               Run using a Consul Key/Value store. This is for distributed installation.
      --consul_key string    Consul Config Key. Can be set using ENV variable. (default "default")
      --consul_uri string    Consul URI. Can be set using ENV variable. (default "localhost:8500")
      --database             Run using a MongoDB database.
  -h, --help                 help for opsilon
      --kubernetes           Run in Kubernetes instead of Docker. You must be connected to a Kubernetes Context
      --local                Run using a local file as config. Not a database. True for CLI. (default true)
      --mongodb_uri string   Mongodb URI. Can be set using ENV variable. (default "mongodb://localhost:27017")

Use "opsilon [command] --help" for more information about a command.
```

# Contribution
Every contribution is welcome. Below is some information to help you get started.

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
