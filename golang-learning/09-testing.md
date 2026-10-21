# 09. Testing

Testing is built into the toolchain: `go test` finds `_test.go` files, runs
functions named `TestXxx`, and reports. No runner to install, no config file.
This chapter covers the `testing` package, the assertion library the repository
uses, the handler test pattern, table tests, and the end to end suite.

## The basics

```go
package api

import "testing"

func TestGetQuoteHandler(t *testing.T) {
	...
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

- File name ends in `_test.go`, function name starts with `Test`, takes
  `*testing.T`.
- `t.Errorf` records a failure and continues; `t.Fatalf` records and stops the
  test. Use `Fatal` when continuing would be meaningless (a nil response you are
  about to dereference).
- `t.Run("name", func(t *testing.T) {...})` creates subtests, which can be run
  selectively.
- `t.Helper()` in a helper function makes failures report the caller's line.
- `t.Cleanup(func)` registers teardown, like `defer` but for the test.

Run them:

```sh
go test ./...                                  # everything
go test ./internal/api                         # one package
go test -run TestGetQuoteHandler ./internal/api # matching name, regex
go test -v ./...                               # print names and logs
go test -race ./...                            # race detector on
go test -cover ./...                           # coverage percentage
go test -count=1 ./...                         # bypass the test cache
```

`go test` caches passing results keyed on inputs. If you change nothing and run
again it prints `(cached)`. `-count=1` forces a rerun.

## quicktest

The repository uses [quicktest](https://github.com/frankban/quicktest) for
assertions, imported as `qt`:

```go
c := qt.New(t)
c.Assert(err, qt.IsNil)
c.Assert(response, qt.Not(qt.IsNil))
c.Assert(response.Quote, qt.Equals, "foo")
c.Assert(err.Error(), qt.Equals, "twirp error invalid_argument: userSession.userUrn invalid user urn")
c.Assert(f.Milliseconds() > int64(500), qt.IsTrue)
```

`Assert` stops the test on failure (like `Fatal`), `Check` continues (like
`Error`). Other useful checkers: `qt.DeepEquals`, `qt.ErrorMatches`,
`qt.Contains`, `qt.HasLen`, `qt.ErrorIs`. Failure output shows both values
nicely formatted, which is the main reason to use it over hand-written `if`s.

The standard library alternative is plain `if got != want { t.Errorf(...) }`.
Plenty of Go code uses only that. Either is fine; be consistent within a
repository.

## The handler test pattern

`internal/api/quote_test.go`:

```go
func TestGetQuoteHandler(t *testing.T) {
	c := qt.New(t)

	// Build a QuoteService backed by a single known quote
	quoteService := quoteservice.NewQuoteService(&quoterepo.QuoteRepository{
		Quotes: []string{"foo"},
	})

	// Create a mock ServerContext
	serverContext := &appserver.ServerContext{
		QuoteService: quoteService,
	}

	// Create a request
	request := &proto.GetQuoteRequest{
		UserSession: &session.UserSession{
			UserUrn:  &wrapperspb.StringValue{Value: "myapp:users:2"},
			Features: []string{},
		},
	}

	// Call the handler
	response, err := GetQuoteHandler(context.Background(), serverContext, request)

	// Assert the results
	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Not(qt.IsNil))
	c.Assert(response.Quote, qt.Equals, "foo")
}
```

Notice what is absent: no HTTP, no generated server, no mocking framework. The
`ServerContext` is a struct, so the test fills in only the field the handler
uses. The repository is real, with data chosen so the outcome is deterministic.
Because dependencies are concrete pointers rather than interfaces, there was
nothing to mock. When a dependency does something you cannot run in a test (a
network call), introduce an interface at that point and write a small fake.

Every RPC has this happy-path test and a second one with an invalid user URN
asserting on the exact error string. Copy both when you add an RPC.

## Table tests

The Go idiom for many cases of one behaviour:

```go
func TestIsValidUrn(t *testing.T) {
	tests := []struct {
		name  string
		urn   string
		types []string
		want  bool
	}{
		{"empty", "", nil, false},
		{"users ok", "myapp:users:1", []string{"users"}, true},
		{"wrong collection", "myapp:tracks:1", []string{"users"}, false},
		{"any collection", "myapp:tracks:1", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(isValidUrn(tt.urn, tt.types), qt.Equals, tt.want)
		})
	}
}
```

An anonymous struct slice, a loop, a subtest per row. The real version is in
`internal/api/validation_test.go`. `go test -run 'TestIsValidUrn/wrong'` runs
one row. `isValidUrn` is unexported but the test is in `package api`, so
it can see it. Tests in the same package can test unexported functions; a
`package api_test` file in the same directory tests only the public surface.

## TestMain and the end to end suite

```go
func TestMain(m *testing.M) {
	...
	if needsLocalServer {
		cmd = runServer(port)
	}
	apiHost = "http://" + host + ":" + port

	exitCode := m.Run()
	if cmd != nil {
		_ = cmd.Process.Kill()
	}
	os.Exit(exitCode)
}
```

`TestMain` runs once per package before any test and is where set-up and
tear-down for the whole package go. `end_to_end/end_to_end_test.go` uses it to
start the compiled binary unless `API_HOST` says to target a deployed instance,
then runs the tests with `m.Run()`, then kills the process.

The tests themselves are black box: they build JSON with `map[string]interface{}`,
POST it, and assert on the status and fields. They deliberately import nothing
from `internal/`, so they prove the wire contract, not the Go types.

`task test:e2e` builds first then runs them. They are excluded from
`task test` by the `grep -v /end_to_end` in the Taskfile because they need a
binary.

## Randomness and time in tests

`GetQuoteFlaky` fails randomly. Its test passes `fail: true` to make the
failure certain rather than trying to assert on a coin flip. `GetQuoteLate`
delays randomly when no delay is given; its test passes an explicit delay and
asserts on elapsed time. The general rule: make nondeterminism injectable, then
inject determinism in tests. If you find yourself wanting to seed the global
random source in a test, pass a `*rand.Rand` in instead.

## Benchmarks and examples

```go
func BenchmarkGimme(b *testing.B) {
	repo := quoterepo.NewQuoteRepository()
	for range b.N {
		_, _ = repo.Gimme(context.Background())
	}
}
```

`go test -bench=. ./internal/quoterepo` runs it and reports ns/op. `Example`
functions with an `// Output:` comment are compiled, run, checked against the
comment, and shown in documentation.

## Try it

1. Run one row of `TestIsValidUrn` by name, then add a row for a URN whose id
   contains a colon and decide what the right answer is.
2. Add `TestGetQuoteFlakyHandlerSucceeds` that... cannot be deterministic as
   written. Change `GetQuoteFlakyHandler` to take its random source from the
   `ServerContext` and write the test.
3. Add a test to `quoterepo` for the empty repository case and assert with
   `qt.ErrorAs` that the error is a `*NoQuotesError`.
4. Write the benchmark above and run it with `-benchmem`.
5. Run `go test -cover ./internal/...` and then `task test:coverage` to see the
   HTML report. Find an untested branch and cover it.

## Read next

- [testing package](https://pkg.go.dev/testing)
- [Go wiki: Table driven tests](https://go.dev/wiki/TableDrivenTests)
- [10. Tooling and code quality](10-tooling.md)
