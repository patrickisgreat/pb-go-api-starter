# 08. Standard library tour

Go's standard library is broad and well designed, and the idiomatic answer to
"which library should I use for X" is often "the one that ships with Go". This
chapter covers the packages the repository uses, in the order you meet them
reading the code.

## `embed`: files inside the binary

```go
import "embed"

//go:embed quotes.json
var f embed.FS

func NewQuoteRepository() *QuoteRepository {
	quotes := &QuoteRepository{}
	data, err := f.ReadFile("quotes.json")
	...
}
```

The `//go:embed` directive above a variable of type `embed.FS`, `string` or
`[]byte` tells the compiler to include the named files in the binary. The path
is relative to the source file and cannot escape the package directory. The
result is one static executable with no data files to ship or find at runtime,
which is why the container image can be `scratch`.

`embed.FS` implements `io/fs.FS`, so you can also serve embedded directories
with `http.FileServer(http.FS(f))` or walk them with `fs.WalkDir`.

## `encoding/json`

```go
_ = json.Unmarshal(data, quotes)
```

`json.Unmarshal(bytes, &value)` decodes into a struct guided by field names and
`json:"..."` tags; `json.Marshal(value)` encodes. Only exported fields
participate. Unknown JSON keys are ignored; missing keys leave the zero value.
Tag options handle the common needs:

```go
Name  string `json:"name"`
Note  string `json:"note,omitempty"`   // omitted when empty
Skip  string `json:"-"`                // never encoded
```

For streams use `json.NewDecoder(r).Decode(&v)` and `json.NewEncoder(w).Encode(v)`,
which the end to end tests could use instead of reading the whole body first.
Decoding into `map[string]interface{}` gives untyped access, with numbers as
`float64`; the end to end tests do this because they intentionally do not import
the generated types.

## `log/slog`: structured logging

```go
log := logger.Get(ctx)
log.Error("no quotes available", slog.Any("err", err))
log.Info("Server started", "addr", sp.Addr)
```

`slog` (Go 1.21+) logs key-value pairs rather than formatted strings, so output
is machine-readable JSON in production and pretty text locally. Attributes can be
loose alternating keys and values (`"addr", sp.Addr`) or typed with
`slog.String`, `slog.Int`, `slog.Any`. `logger.Get(ctx)` in `internal/logger` returns a
`*slog.Logger` that already carries the trace and span IDs from the context, so
log lines correlate with traces without any effort from you.

The older `log` package (`log.Fatalf` in `context.go`) is still fine for
start-up failures where structure does not matter.

## `math/rand`

```go
ix := rand.Intn(len(cr.Quotes))
delay = int32(rand.Float64()*900 + 100)
```

`rand.Intn(n)` gives `[0, n)`, `rand.Float64()` gives `[0, 1)`. Since Go 1.20 the
global source is seeded randomly at start-up, so no `rand.Seed` call is needed.
For anything security related use `crypto/rand` instead. Go 1.22 added
`math/rand/v2` with a cleaner API and `rand.N`; new code can use it.

## `context`, `sync`, `os/signal`, `syscall`

Covered in [chapter 06](06-context-goroutines-shutdown.md).

## `net`, `net/http`

```go
net.JoinHostPort(config.Config.Host, fmt.Sprintf("%d", config.Config.Port))
net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
```

`net.JoinHostPort` exists because IPv6 addresses need brackets; do not
concatenate with `:` yourself. `net/http` is [chapter 07](07-http-protobuf-twirp.md).
The client side used in the end to end tests:

```go
req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
req.Header.Set("Content-Type", "application/json; charset=utf-8")
resp, err := client.Do(req)
defer resp.Body.Close()
body, err := io.ReadAll(resp.Body)
```

Always close the body, always with `defer` right after the error check.
`http.NewRequestWithContext` is the version to prefer in code that has a
context.

## `time`

```go
time.Sleep(time.Duration(delay) * time.Millisecond)
context.WithTimeout(ctx, 10*time.Second)
s := time.Now(); f := time.Since(s); f.Milliseconds()
```

`time.Duration` is an `int64` of nanoseconds with constants `time.Millisecond`,
`time.Second` and so on. `2 * time.Second` is a `Duration`; a bare `2` is not,
so `time.Sleep(2)` sleeps for two nanoseconds. Converting an integer variable
requires the explicit `time.Duration(n)` cast you see in `late.go`.

`time.Time` values are for instants; format them with layouts based on the
reference time `2006-01-02 15:04:05` or use `time.RFC3339`.

## `os`, `os/exec`, `path/filepath`

```go
here, err := os.Getwd()
serverPath, err := filepath.Abs(filepath.Join(here, "../cmd/app/bin/pb-go-api-starter"))
cmd := exec.Command(serverPath)
err = cmd.Start()
_ = cmd.Process.Kill()
os.Getenv("API_HOST")
os.Exit(exitCode)
```

The end to end `TestMain` spawns the compiled server as a subprocess. `filepath`
is for OS paths, `path` is for URL-style slash paths; use the right one.
`os.Getenv` returns `""` for unset variables; `os.LookupEnv` tells you whether it
was set.

## `fmt`, `errors`, `strings`, `slices`

`fmt.Sprintf`, `fmt.Errorf`, `fmt.Fprint(w, "OK")`: formatting to strings,
errors and writers. `errors.New`, `errors.Is`, `errors.As`: [chapter 04](04-errors.md).

`strings` has everything you expect: `Split`, `Join`, `TrimSpace`, `HasPrefix`,
`Contains`, `ToLower`, `Builder`. `slices` (Go 1.21+) adds generic helpers:

```go
slices.Contains(urnTypes, u.Collection)    // in validation.go
slices.Sort(s)
slices.Index(s, v)
```

`maps` is the map equivalent (`maps.Keys`, `maps.Clone`).

## `testing`

[Chapter 09](09-testing.md).

## Finding things

```sh
go doc encoding/json            # package overview in the terminal
go doc json.Unmarshal           # one function
go doc -all strings | less      # everything
```

[pkg.go.dev](https://pkg.go.dev/std) is the same documentation online, with
runnable examples. Every package page starts with an overview worth reading once.

## Try it

1. Add a second embedded file, `authors.json`, and expose it through a new
   repository method. Keep the parsing in `NewQuoteRepository`.
2. Change `NewQuoteRepository` to return an error when the JSON is invalid,
   and make `InitContext` fail loudly if it does. Corrupt the file to test it.
3. Replace the `map[string]interface{}` in the end to end tests with a small
   struct with `json` tags, and compare the readability.
4. Write a `MarshalJSON` method on a type and see `encoding/json` pick it up.
5. Use `slog.With("component", "quoterepo")` to create a child logger with a
   fixed attribute and log through it.

## Read next

- [Standard library index](https://pkg.go.dev/std)
- [Go blog: Structured Logging with slog](https://go.dev/blog/slog)
- [09. Testing](09-testing.md)
