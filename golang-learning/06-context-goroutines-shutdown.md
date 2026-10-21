# 06. Context, goroutines and shutdown

Concurrency is built into the language: goroutines are cheap threads managed by
the runtime, channels move values between them, and `context.Context` carries
cancellation and deadlines through call chains. The server in
`internal/appserver/server.go` uses all three in thirty lines, so it is the
worked example for this chapter.

## Goroutines

```go
go func() {
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("Failed to start server", slog.Any("err", err))
		panic(err)
	}
}()
```

`go f()` starts `f` on a new goroutine and returns immediately. Goroutines are
multiplexed onto OS threads by the runtime; starting a hundred thousand is fine.
Here the HTTP server, which blocks forever in `ListenAndServe`, runs in the
background so that `Start` can go on to wait for a shutdown signal.

The function literal `func() { ... }()` is a closure, declared and called in one
step. It captures `server` and `log` from the enclosing scope.

The `net/http` server itself starts a goroutine per connection. Your handlers
already run concurrently, which means anything they share must be safe to use
from many goroutines at once. `QuoteRepository` is only read after start-up, so
it is safe. A map that handlers write to would not be.

## Channels

```go
stop := make(chan os.Signal, 1)
signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

// Wait for SIGINT or SIGTERM
<-stop
```

A channel is a typed pipe. `make(chan T, n)` creates one with buffer size `n`;
`ch <- v` sends, `v := <-ch` receives, and both block when they cannot proceed.
`signal.Notify` asks the runtime to send OS signals into the channel, and
`<-stop` blocks the goroutine until one arrives. That single line is the whole
"run until told to stop" logic.

The buffer of 1 matters: `signal.Notify` does not block when delivering, so
without a buffer a signal that arrives before the receive would be dropped.

Channels compose with `select`, which waits on several at once:

```go
select {
case sig := <-stop:
	log.Info("signal", "sig", sig)
case <-ctx.Done():
	log.Info("context cancelled")
case <-time.After(time.Minute):
	log.Info("timed out")
}
```

You will not need `select` in this repository yet, but it is the idiom for
timeouts and for combining cancellation with work.

## `context.Context`

Almost every function in the repository takes `ctx context.Context` as its
first parameter, and passes it on. A context carries three things down a call
chain:

1. **Cancellation.** `ctx.Done()` is a channel that closes when the context is
   cancelled. Long operations check it or pass it to libraries that do.
2. **Deadlines.** `context.WithTimeout` and `context.WithDeadline` create a child
   context that cancels itself at a time.
3. **Values.** Request-scoped data such as the logger and trace span.
   `logger.Get(ctx)` in `internal/logger` pulls the logger out of the context.

Shutdown uses a timeout:

```go
ctxWithTimeout, cancel := context.WithTimeout(ctx, TEN_SECONDS)
defer cancel()

if err := server.Shutdown(ctxWithTimeout); err != nil {
	log.Error("Failed to shut down server", slog.Any("err", err))
}
```

`Shutdown` stops accepting connections and waits for in-flight requests, but
gives up when `ctxWithTimeout` expires. The `defer cancel()` releases the
timer if `Shutdown` finishes early; always call the cancel function you get
back, even when you do not think you need to (`go vet` will remind you).

Contexts form a tree. `context.Background()` in `main` is the root; every
request gets a child created by the HTTP server, cancelled when the client
disconnects; handlers derive further children for timeouts. Cancelling a parent
cancels every descendant.

Rules:

- `ctx` is always the first parameter and is never stored in a struct.
- Never pass `nil`; use `context.Background()` at the top level and
  `context.TODO()` when you have not decided yet.
- Use context values for request-scoped metadata only, never for optional
  function parameters.

## Once-only initialisation

```go
var initOnce sync.Once

func InitContext(ctx context.Context) {
	initOnce.Do(func() {
		...
		sctx = newContext()
	})
}
```

`sync.Once.Do` runs its function the first time and never again, and it is safe
to call from multiple goroutines at once: the second caller blocks until the
first finishes. This is how `appserver` guarantees one `ServerContext` even if
`InitContext` is called twice.

The `sync` package also has `Mutex` and `RWMutex` for guarding shared state,
`WaitGroup` for waiting on a set of goroutines, and `atomic` for lock-free
counters. The pattern for a shared map is:

```go
type cache struct {
	mu sync.RWMutex
	m  map[string]string
}

func (c *cache) get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[k]
	return v, ok
}
```

## The race detector

`go test -race ./...` instruments the binary to detect unsynchronised concurrent
access at runtime. It is slow but it finds real bugs. Run it whenever you add
shared state.

## The full lifecycle

Read `Start` in `internal/appserver/server.go` top to bottom with this chapter
in mind:

1. Refuse to start if the context was never initialised (a programming error,
   so panic).
2. Build an `http.Server`.
3. Start it in a goroutine.
4. Block on the signal channel.
5. On signal, shut down with a ten second budget, then flush telemetry.

Every deployed Go service has some version of this function.

## Try it

1. Write a program that starts ten goroutines, each sending its index on a
   channel, and prints the ten values. Then use a `sync.WaitGroup` to wait for
   them instead.
2. Add a per-request timeout to `GetQuoteLateHandler`: derive
   `context.WithTimeout(ctx, 3*time.Second)` and return `twirp.DeadlineExceeded`
   if the delay would exceed it. Hint: replace `time.Sleep` with a `select` on
   `time.After` and `ctx.Done()`.
3. Add a request counter to `ServerContext`, increment it in each handler
   without a lock, and run `go test -race ./...` with a test that calls the
   handler from two goroutines. Then fix it with `sync/atomic`.
4. Send `SIGTERM` to the running server (`kill -TERM <pid>`) while a
   `GetQuoteLate` request with a long delay is in flight and watch the logs.

## Read next

- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency)
- [07. HTTP, protobuf and Twirp](07-http-protobuf-twirp.md)
