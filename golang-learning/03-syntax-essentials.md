# 03. Syntax essentials

Enough Go to read every file in this repository without stopping. Each section
shows the construct, then where it appears in the code.

## Declarations

```go
var host string            // zero value: ""
var port = 8000            // type inferred: int
count := 0                 // short declaration, only inside functions
x, y := 1, "two"           // multiple at once
const TEN_SECONDS time.Duration = 10 * time.Second
```

`:=` declares and assigns; `=` only assigns. You will use `:=` almost everywhere
inside functions. Every variable has a zero value: `0`, `""`, `false`, `nil` for
pointers, slices, maps, channels, interfaces and functions. There is no
`undefined` and no uninitialised memory.

From `internal/appserver/server.go`:

```go
const TEN_SECONDS time.Duration = 10 * time.Second
```

## Types

```go
int, int32, int64, uint8 (byte), float64, string, bool, rune (int32, a code point)
[]string          slice of strings
[5]int            array, fixed size, rarely used directly
map[string]int    map
*T                pointer to T
func(int) error   function type
struct { ... }    struct
interface { ... } interface
chan T            channel
```

Types come after names: `func f(ctx context.Context, delay int32) (string, error)`.
Read it left to right: "f takes ctx and delay and returns a string and an error".

There is no implicit numeric conversion. `int32(delay)` and `time.Duration(delay)`
are explicit casts, and `internal/api/late.go` has both:

```go
delay = int32(rand.Float64()*900 + 100)
time.Sleep(time.Duration(delay) * time.Millisecond)
```

## Functions and multiple returns

```go
func (cr *QuoteRepository) Gimme(ctx context.Context) (string, error) {
	...
	return "", err
	...
	return cr.Quotes[ix], nil
}
```

Functions return multiple values, and the last one is conventionally an `error`.
The caller must handle every returned value; use `_` to discard one on purpose:

```go
quote, err := quoteService.GetQuote(ctx)
_ = json.Unmarshal(data, quotes)        // ignoring the error deliberately
```

Named results are allowed but the repository does not use them. Variadic
parameters use `...`: `func NewQuotesServer(svc Quotes, opts ...interface{})`.

## Structs and pointers

```go
type StartParams struct {
	Addr    string
	Handler http.Handler
}

appserver.Start(ctx, &appserver.StartParams{
	Handler: telemetry.InstrumentHandler("Quotes", router),
	Addr:    net.JoinHostPort(config.Config.Host, fmt.Sprintf("%d", config.Config.Port)),
})
```

Struct literals name their fields. `&T{...}` creates the struct and takes its
address, giving a `*T`. Go dereferences pointers to structs for you when
accessing fields and calling methods, so `sp.Addr` works whether `sp` is a
`StartParams` or a `*StartParams`.

When to use a pointer:

- The function needs to modify the struct.
- The struct is large and copying it would be wasteful.
- The type has methods with pointer receivers (chapter 05).
- You need `nil` to mean "absent".

`ServerContext`, `QuoteRepository` and `QuoteService` are all passed around as
pointers so that every part of the program shares one instance.

There is no pointer arithmetic, and `nil` pointer dereference panics rather than
corrupting memory.

## Slices and maps

```go
Quotes []string                          // a slice: pointer to array + length + capacity
ix := rand.Intn(len(cr.Quotes))          // len works on slices, maps, strings, channels
cr.Quotes[ix]                            // index; out of range panics
[]string{"foo"}                          // slice literal
append(s, "bar")                         // returns a new slice; assign it back
s[1:3]                                   // sub-slice, shares the backing array

payload := map[string]interface{}{       // map literal
	"fail": true,
}
v, ok := payload["fail"]                 // comma-ok: ok is false if the key is absent
delete(payload, "fail")
```

A slice is a view over an array. Two slices can share memory, and `append` may
or may not allocate a new array, which is why you always write
`s = append(s, x)`. Maps are reference types; a `nil` map can be read but
writing to it panics, so create them with `make(map[K]V)` or a literal.

Neither slices nor maps are safe for concurrent writes. Chapter 06 covers that.

## Control flow

```go
if err != nil {
	return nil, err
}

if request.GetFail() || rand.Intn(2) == 0 {
	return nil, twirp.InternalError("flaky error")
}

for retries < 10 && !isPortOpen("127.0.0.1", port) {   // the only loop keyword
	time.Sleep(2 * time.Second)
}

for i, q := range cr.Quotes { ... }      // index and value
for _, q := range cr.Quotes { ... }      // value only
for range 3 { ... }                      // three times (Go 1.22+)
for { ... }                              // forever; break to exit

switch code {
case 200, 204:
	...
case 500:
	...
default:
	...
}
```

No parentheses around conditions, braces required, opening brace on the same
line (the formatter enforces this and the grammar demands it). `switch` cases do
not fall through unless you say `fallthrough`. `if` can start with a statement:
`if err := f(); err != nil { ... }` scopes `err` to the `if`.

`internal/appserver/server.go` combines `if` and `:=` this way:

```go
if err := server.Shutdown(ctxWithTimeout); err != nil {
	log.Error("Failed to shut down server", slog.Any("err", err))
}
```

## `defer`

```go
ctxWithTimeout, cancel := context.WithTimeout(ctx, TEN_SECONDS)
defer cancel()

resp, err := client.Do(req)
if err != nil {
	return nil, 0, err
}
defer resp.Body.Close()
```

`defer` schedules a call to run when the enclosing function returns, however it
returns. It is Go's `finally`, `using` and `with` rolled into one, and the
idiom is to write the `defer` on the line immediately after acquiring the thing
it releases. Deferred calls run last-in, first-out.

## Strings

Strings are immutable byte sequences, almost always UTF-8. `len(s)` is bytes,
not characters. `range` over a string yields runes. Formatting is `fmt.Sprintf`
with `%d`, `%s`, `%v` (anything), `%+v` (struct with field names), `%q` (quoted),
`%w` (wrap an error, chapter 04). Concatenate with `+`; build big strings with
`strings.Builder`.

## `interface{}` and `any`

`interface{}` is the empty interface, satisfied by every type. `any` is its alias
since Go 1.18. `map[string]interface{}` in the end to end tests is "a JSON object
of unknown shape". Get a concrete value back out with a type assertion:

```go
msg, ok := response["msg"].(string)
```

Generics exist (`func Load[T any](cfg *T) error` in `internal/baseconfig`) but you will rarely
need to write them. Reading them is enough for now.

## What is not there

No classes, no inheritance, no constructors (a function named `NewX` is the
convention), no exceptions, no ternary operator, no function overloading, no
default parameter values, no enums (typed constants with `iota` instead), no
implicit conversions, no unused variables or imports, no semicolons that you
have to type.

## Try it

1. In `internal/quoterepo/repository.go`, write `func (cr *QuoteRepository) Longest() string`
   that returns the longest quote using `range` and `len`.
2. Write a function that takes `[]string` and returns `map[string]int` counting
   how many quotes start with each first letter. Use the comma-ok form to test
   for a key.
3. Add a `defer fmt.Println("done")` at the top of `Gimme` and call it. Observe
   when it prints relative to the `return`.
4. Try compiling with an unused variable. Then try `for i := 0; i < 3; i++ {}`
   and check that the classic three-clause form works too.

## Read next

- [A Tour of Go](https://go.dev/tour/basics/1): skim Basics and Flow control.
  You now know enough to go fast.
- [04. Errors](04-errors.md)
