# 01. Toolchain and modules

Go ships as one binary, `go`, that builds, tests, formats, vets, fetches
dependencies and manages versions. There is no separate package manager, build
tool or formatter to choose. This chapter shows the commands you will use daily
and what the two files at the root of the repository, `go.mod` and `go.sum`, are.

## The commands that matter

```sh
go version              # which toolchain you have
go build ./...          # compile every package under the current directory
go run ./cmd/app        # compile and run a main package in one step
go test ./...           # run every test
go vet ./...            # static checks the compiler does not do
go fmt ./...            # rewrite files in canonical formatting
go mod tidy             # add missing and remove unused dependencies
go get -u ./...         # upgrade dependencies
go install pkg@version  # build a tool and put it in $(go env GOPATH)/bin
go env GOPATH GOOS      # print toolchain settings
```

`./...` is a pattern meaning "this directory and everything below it". You will
type it constantly.

`go build ./...` compiles everything and throws the result away. It is the
fastest way to check that the whole repository still compiles. To get an actual
binary you name a `main` package and an output path, which is what the Taskfile
does:

```sh
CGO_ENABLED=0 GOOS=darwin go build -v -o ./bin/pb-go-api-starter   # from cmd/app
```

`CGO_ENABLED=0` disables C interop so that the binary is fully static and runs on
a `scratch` container image. `GOOS` and `GOARCH` cross-compile: set
`GOOS=linux GOARCH=arm64` and you get a Linux binary from a Mac with no extra
setup. Look at `Dockerfile` to see it used.

## Modules

A module is a tree of packages with a `go.mod` file at its root. This repository
is one module. Open `go.mod`:

```
module github.com/patrickisgreat/pb-go-api-starter

go 1.26.0

require (
	github.com/bakins/twirpotel v0.2.0
	github.com/frankban/quicktest v1.14.6
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/twitchtv/twirp v8.1.3+incompatible
	go.opentelemetry.io/contrib/exporters/autoexport v0.72.0
	...
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	...
)
```

- **`module`** is the module path. It is also the import prefix for every package
  inside: `internal/api` is imported as
  `github.com/patrickisgreat/pb-go-api-starter/internal/api`. Module paths look
  like URLs because `go get` uses them to find the source, but nothing is fetched
  for your own module.
- **`go 1.26.0`** is the minimum language version. Newer toolchains build it fine.
- **`require`** lists dependencies with exact versions. The second block, marked
  `// indirect`, is dependencies of your dependencies. You never edit these blocks
  by hand; `go get` and `go mod tidy` do.

`go.sum` holds cryptographic hashes for every module version ever used, so a
later download that does not match is rejected. Commit it. Do not edit it.

### Versions

Go uses semantic versioning and enforces it in the import path: a module at
major version 2 or later must have `/v2` at the end of its path, which is why you
see `github.com/cespare/xxhash/v2`. `+incompatible` on twirp means that module
predates this rule. `v0.0.0-20221221133751-67e37ae746cd` is a pseudo-version
pointing at a commit that was never tagged.

### Where dependencies live

There is no `node_modules`. Downloaded modules go to a shared, read-only cache
at `$(go env GOMODCACHE)` (usually `~/go/pkg/mod`), keyed by version, so every
project on your machine shares one copy of each version.

By default `go` fetches through a public proxy and checks a public checksum
database. Every dependency of this repository is public, so that just works. If
you ever depend on a private module, `GOPRIVATE=github.com/yourorg/*` tells Go
to bypass the proxy and checksum database for it and clone with `git` instead.

### Adding a dependency

Import it in code, then run `go mod tidy`. Go resolves the latest version, adds it
to `go.mod` and `go.sum`, and downloads it. To pin or upgrade a specific version:

```sh
go get github.com/some/module@v1.4.2
go get github.com/some/module@latest
```

### The `internal` rule

Any package under a directory named `internal` can only be imported by code
rooted at the parent of that `internal`. For this module, everything under
`internal/` is importable from anywhere in the module and from nowhere else.
This is enforced by the compiler, not a linter. Chapter 02 covers how the
repository uses it.

## GOPATH is history

You will find old tutorials that talk about `GOPATH` and putting your code in
`~/go/src/...`. Ignore them. Since Go 1.13 modules replaced that workflow. The
only thing `GOPATH` is used for today is the location of the module cache and of
binaries installed with `go install`, which is why `$(go env GOPATH)/bin` needs
to be on your `PATH`.

## Task, air and the rest

The repository wraps the `go` commands in a `Taskfile.yml` so that flags and
environment variables live in one place. `task build` runs code generation first,
`task test` runs tidy, lint and vet first. Read `Taskfile.yml` top to bottom once;
it is the map of every workflow in the repository.

`air` is a file watcher that reruns `task build` and restarts the binary on save.
It is not part of Go; it is a convenience for local development.

## Try it

1. Run `go build ./...` from the repository root. Then delete
   `internal/generated` and run it again. Read the error, then run `task gen:rpc`
   and build once more. (`git status` will show the files restored unchanged,
   because the generated code is committed.)
2. Run `go env` and find `GOMODCACHE`. List that directory and find the twirp
   module.
3. Cross-compile: from `cmd/app` run
   `GOOS=linux GOARCH=amd64 go build -o /tmp/pb-linux .` and then `file /tmp/pb-linux`.
4. Run `go mod why github.com/frankban/quicktest` to see which package pulls in
   the test assertion library.

## Read next

- [Go Modules Reference](https://go.dev/ref/mod), sections "Introduction" and
  "go.mod files".
- [02. Packages and project layout](02-packages-and-layout.md)
