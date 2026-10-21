# Testing

There are two kinds of tests. Unit tests exercise a package in isolation; end to
end tests start the real binary and talk to it over HTTP.

## Unit tests

```sh
task test            # tidy, lint, vet, then go test on every package except rpc and end_to_end
task test:ci         # the same without the formatting steps, for CI
task test:coverage   # writes coverage.out and opens the HTML report
```

Tests live beside the code in `_test.go` files and use
[quicktest](https://github.com/frankban/quicktest) for assertions:

```go
c := qt.New(t)
c.Assert(err, qt.IsNil)
c.Assert(response.Quote, qt.Equals, "foo")
```

### Testing a handler

Handlers take a `*appserver.ServerContext`, which is a plain struct. Tests build
one by hand with a repository holding a single known quote, so the outcome is
deterministic and no server, network or generated Twirp code is involved:

```go
quoteService := quoteservice.NewQuoteService(&quoterepo.QuoteRepository{
    Quotes: []string{"foo"},
})
serverContext := &appserver.ServerContext{QuoteService: quoteService}

response, err := GetQuoteHandler(context.Background(), serverContext, request)
```

Each RPC has a happy-path test and an invalid-session test. When you add an RPC,
add both.

### Randomness

`GetQuoteFlaky` fails randomly and `GetQuoteLate` delays randomly when no delay is
given. Their tests pin the behaviour instead: `fail: true` for a guaranteed
failure, and an explicit `delayMs` for the late handler.

## End to end tests

```sh
task test:e2e        # build the binary, then run end_to_end/
task test:e2e:ci     # run end_to_end/ against an already running instance
```

`end_to_end/end_to_end_test.go` has a `TestMain` that decides where to send
requests:

- If `API_HOST` is unset it starts `cmd/app/bin/pb-go-api-starter` on
  `127.0.0.1:8000`, waits for the port to open, runs the tests, and kills the
  process.
- If `API_HOST` is set it targets `http://$API_HOST:$API_PORT` directly. The
  `.env/.env.development`, `.env.staging` and `.env.production` files set these,
  so `ENV=staging task test:e2e:ci` runs the suite against staging.

The tests themselves in `end_to_end/quote_test.go` are black box: they `POST` JSON
to the Twirp route and assert on status code and response fields. They do not
import any package from `internal/`.

## Linting and vetting

`task lint` runs `golangci-lint` with the configuration in `.golangci.yml`. The
only non-default linter is `goimports` with `github.com/patrickisgreat/pb-go-api-starter`
as the local prefix, which keeps imports grouped as standard library, third party,
then this module. `task format` fixes import grouping automatically.

`task vet` runs `go vet`, which catches things the compiler allows but that are
almost always bugs, such as printf format mismatches and unreachable code.

CI runs all of the above plus both test suites under `go test -race`. If a test
passes locally but fails in CI with a data race report, the race is real; the
detector only reports accesses that actually happened.

## What is not tested

`cmd/app/main.go`, `internal/appserver/server.go` and `internal/appserver/context.go`
have no unit tests. They wire real dependencies together and are covered by the
end to end suite, which cannot pass unless the server starts, serves and shuts
down.
