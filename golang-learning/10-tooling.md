# 10. Tooling and code quality

Go settles most style arguments by tooling. This chapter covers the tools the
repository runs, what each catches, and how to wire them into your editor so
that `task lint` never surprises you.

## `gofmt` and `goimports`

`gofmt` rewrites source into the one canonical layout: tabs, brace placement,
spacing, alignment of struct fields and comments. There are no options. Every Go
codebase in the world looks the same, and code review never discusses
formatting.

`goimports` is `gofmt` plus import management: it adds missing imports, removes
unused ones, and groups them. The repository runs it with a local prefix so that
this module's packages form their own group:

```sh
goimports -w -local "github.com/patrickisgreat/pb-go-api-starter" .
```

That is `task format`. Configure your editor to run `goimports` on save
(`gopls`, the Go language server, does this out of the box in VS Code, GoLand,
Neovim and others), and you will never run it by hand.

## `go vet`

```sh
go vet ./...
```

Static analysis for mistakes that compile but are almost certainly wrong:
`Printf` verbs that do not match arguments, unreachable code, copying a
`sync.Mutex` by value, `context.WithCancel` results whose cancel function is
never called, struct tags with typos, and more. It is fast and has almost no
false positives, so `task test` runs it before every test run. Treat a vet
failure as a bug.

## `golangci-lint`

A runner for dozens of linters with one configuration file. `.golangci.yml`:

```yaml
version: "2"

linters:
  exclusions:
    paths:
      - rpc

formatters:
  enable:
    - goimports
  settings:
    goimports:
      local-prefixes:
        - github.com/patrickisgreat/pb-go-api-starter
```

The default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`,
and others) is on. Version 2 of the tool separates formatters from linters, so
`goimports` is listed under `formatters` and `golangci-lint fmt` applies it.
`errcheck` is the one you will meet most: it flags any error return value that
is not handled, which is why the repository writes `_ = json.Unmarshal(...)` to
mark a deliberate ignore rather than just dropping the result.

Run with `task lint`. To see what a linter is complaining about, the message
includes its name in parentheses; `golangci-lint help linters` describes each.
Add linters as the team agrees on them; `gocritic`, `revive` and `gosec` are
common next steps.

## `go mod tidy`

Keeps `go.mod` and `go.sum` exactly in sync with what the code imports. Run it
after adding or removing an import; `task test` runs it first. A CI check that
`go mod tidy` produces no diff is a common guard against stale dependency files.

## `gopls`

The official language server. It powers completion, go-to-definition,
rename-across-packages, find-references, inline diagnostics from `vet` and
`staticcheck`, and format-on-save. Any editor that speaks LSP gets all of it.
If your editor shows red squiggles that the command line does not, `gopls` is
usually ahead of you: it runs analysers the compiler does not.

Refactoring by hand is rarely needed. Rename a symbol with `gopls` and every
reference in the module follows.

## `air`

Not a Go tool, but the reason `task run` feels like a scripting language. It
watches the directories listed in `.air.toml`, runs `task build` on change, and
restarts the binary. Build errors land in `build-errors.log`. Tests and
`internal/generated` are excluded from the watch so that regeneration does not
trigger a rebuild loop.

## `go doc` and `pkg.go.dev`

```sh
go doc net/http.Handler
go doc -all github.com/twitchtv/twirp | less
```

Doc comments are ordinary comments directly above a declaration, starting with
the name:

```go
// NewQuoteRepository loads the embedded quotes and returns a repository
// ready for use. An unreadable or invalid file yields an empty repository.
func NewQuoteRepository() *QuoteRepository {
```

`golangci-lint` can enforce these on exported identifiers (`revive`'s
`exported` rule). The repository does not yet; adding doc comments to the
exported functions in `internal/` is a good first contribution.

## Build tags and generate

`//go:build integration` at the top of a file includes it only when
`go build -tags integration` is used; some teams gate slow tests this way.
`//go:generate protoc ...` comments let `go generate ./...` run code generators,
an alternative to the `task gen:rpc` task. Neither is used here, but you will
see both.

## Continuous integration

A minimal Go CI job is:

```sh
task gen:rpc
go mod tidy && git diff --exit-code go.mod go.sum
task lint
task test:ci
task build
task test:e2e:ci
```

Everything in this chapter runs in seconds, so there is no reason to skip any of
it on every push.

## Try it

1. Break the formatting of a file (put a brace on its own line) and run
   `gofmt -d` on it to see the diff, then `task format`.
2. Call `os.Open` somewhere without checking the error and run `task lint`.
   Then fix it two ways: handle it, and explicitly discard it.
3. Write `fmt.Printf("%d", "hello")` and run `go vet`.
4. Enable `revive` in `.golangci.yml`, run the linter, and add doc comments
   until it is quiet.
5. Set up format-on-save in your editor and confirm imports are grouped
   correctly when you add one by hand.

## Read next

- [gofmt](https://pkg.go.dev/cmd/gofmt), [go vet](https://pkg.go.dev/cmd/vet)
- [golangci-lint linters](https://golangci-lint.run/usage/linters/)
- [11. Capstone: add an RPC](11-capstone-add-an-rpc.md)
