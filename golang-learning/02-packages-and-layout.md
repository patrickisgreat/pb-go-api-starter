# 02. Packages and project layout

A Go program is a set of packages. A package is a directory of `.go` files that
all start with the same `package` line. This chapter explains how packages work,
what the repository's directories mean, and the one visibility rule that replaces
`public`, `private` and `protected`.

## One directory, one package

Every `.go` file in `internal/api` begins with `package api`. The compiler treats
the directory as a unit: files in the same package share a namespace and can use
each other's unexported names without importing anything. `validation.go`
defines `validateUserSession`, and `quote.go` calls it as if it were in the same
file, because from Go's point of view it is.

Rules that follow from this:

- A directory holds exactly one package (plus optional `_test` variants, see
  chapter 09).
- The package name is usually the last element of the directory path:
  `internal/quoterepo` is `package quoterepo`.
- Package names are short, lower case, no underscores or camel case. `quoterepo`
  rather than `quote_repository` or `QuoteRepo`.
- `package main` is special: it produces an executable and must contain `func main()`.

## Imports

```go
import (
	"context"
	"fmt"
	"net"

	"github.com/twitchtv/twirp"

	"github.com/patrickisgreat/pb-go-api-starter/internal/appserver"
	proto "github.com/patrickisgreat/pb-go-api-starter/internal/generated/pb_go_api_starter/api"
)
```

This is `internal/api/flaky.go`. Three things to notice:

1. Imports are full paths. Standard library packages have short paths
   (`context`, `net/http`); everything else starts with the module path.
2. The blank lines separate groups: standard library, third party, this module.
   `goimports` maintains this ordering automatically (chapter 10).
3. `proto "..."` gives the package a local alias. The generated package is named
   `api`, which would collide with the importing package `api`, so it is
   imported as `proto`. Without an alias you refer to a package by its declared
   name, which is the last path element: `appserver.Start`, `twirp.InternalError`.

Unused imports are a compile error, not a warning. So are unused local variables.
Your editor's Go plugin will add and remove imports as you type; let it.

## Exported means capitalised

There are no visibility keywords. An identifier that starts with an upper case
letter is exported from its package; one that starts lower case is not.

```go
type QuoteRepository struct {      // exported type
	Quotes []string                // exported field
}

func NewQuoteRepository() *QuoteRepository   // exported constructor

var f embed.FS                     // unexported package-level variable
```

This applies to types, functions, methods, struct fields, constants and
variables. It is why JSON field names need struct tags (chapter 08): the JSON
library can only see exported fields, and exported fields are capitalised.

Within a package everything is visible, so "private to the package" is the only
privacy level. That is coarser than a class, and it is why packages are kept
small and focused.

## Package-level state and `init`

A package can declare variables at the top level:

```go
var sctx *ServerContext
var shutdownOtel func(context.Context) error
var initOnce sync.Once
```

These live for the life of the process. `internal/appserver/context.go` uses
them to hold the single `ServerContext`. The convention is to guard such
initialisation with `sync.Once` or an explicit `Init` function called from
`main`, which is what `appserver.InitContext` is. Go also has `func init()`
which runs automatically when a package is loaded, but it makes initialisation
order hard to follow and this repository does not use it.

## The repository layout

```
cmd/app/            package main
internal/api        package api
internal/appserver  package appserver
internal/baseconfig package baseconfig
internal/config     package config
internal/handlers   package handlers
internal/logger     package logger
internal/quoterepo  package quoterepo
internal/quoteservice package quoteservice
internal/telemetry  package telemetry
internal/urn        package urn
internal/generated  packages written by protoc; committed, never hand-edited
rpc/                .proto files, not Go
end_to_end/         package end_to_end, test only
```

### `cmd/`

Executables. One subdirectory per binary, each a `package main`. Keeping `main`
tiny (`cmd/app/main.go` is 30 lines) means nothing important lives in a package
that cannot be imported or tested.

### `internal/`

Private to this module. The compiler refuses to let another module import
anything below an `internal` directory. Since this is a service and not a library,
almost all code goes here. If you later extract something for other modules to
use, it moves to a `pkg/` directory or its own module.

### Dependency direction

```
main → handlers → api → quoteservice → quoterepo
  ↘ appserver ↗    ↘ urn              ↘ logger
      ↘ config, logger, telemetry
```

`appserver` builds the `ServerContext` from `quoterepo` and `quoteservice`, and
`api` receives that context. `baseconfig`, `logger`, `telemetry` and `urn` sit
at the bottom: they import nothing else from the repository, so anything can
import them. Go forbids import cycles outright: if `quoterepo`
tried to import `api`, the build would fail. Layouts like this one are shaped as
much by that rule as by taste.

### Test files

Files ending in `_test.go` are compiled only by `go test`. They sit next to the
code they test in the same directory and, usually, the same package. `go build`
ignores them.

## Try it

1. Create `internal/quoterepo/count.go` with a function `func (cr *QuoteRepository) Count() int`.
   Call it from `internal/api/quote.go` and log the result. Notice you needed no
   import for the method itself because the type was already imported.
2. Rename `Count` to `count` and build. Read the error.
3. Add `import "github.com/patrickisgreat/pb-go-api-starter/internal/api"` to
   `internal/quoterepo/repository.go` and build. Read the import cycle error, then
   remove it.
4. Run `go list ./...` to see every package path in the module.

## Read next

- [Effective Go: Names](https://go.dev/doc/effective_go#names)
- [03. Syntax essentials](03-syntax-essentials.md)
