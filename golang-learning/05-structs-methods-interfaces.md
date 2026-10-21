# 05. Structs, methods and interfaces

Go has no classes. It has structs for data, methods attached to any named type,
and interfaces that are satisfied implicitly. Together they cover everything you
used classes and inheritance for, with less ceremony and no hierarchy.

## Methods

A method is a function with a receiver, declared between `func` and the name:

```go
type QuoteService struct {
	repo *quoterepo.QuoteRepository
}

func (cs *QuoteService) GetQuote(ctx context.Context) (string, twirp.Error) {
	quote, err := cs.repo.Gimme(ctx)
	...
}
```

`cs` is the receiver, playing the role of `this` or `self`, and you name it
yourself, conventionally one or two letters. Methods can be declared on any type
you define in the package, not only structs. `NoQuotesError` has a method, and
so could `type Celsius float64`.

### Pointer versus value receivers

`(cs *QuoteService)` is a pointer receiver: the method sees and can modify the
caller's struct. `(c Celsius)` would be a value receiver: the method gets a
copy. Rules:

- If any method needs to mutate the receiver, or the struct is big, use pointers.
- Be consistent: all pointer or all value receivers on a given type.
- Go takes the address for you: `svc.GetQuote(ctx)` works whether `svc` is a
  `QuoteService` or a `*QuoteService`, as long as `svc` is addressable.

Every type in this repository uses pointer receivers, which is the common default
for anything that holds state.

## Constructors

There are none. The convention is a function `NewT` returning `*T` or `T`:

```go
func NewQuoteService(repo *quoterepo.QuoteRepository) *QuoteService {
	return &QuoteService{
		repo: repo,
	}
}
```

Because `repo` is unexported, code outside `quoteservice` cannot build a
`QuoteService` any other way, so the constructor is the only door. Compare
`QuoteRepository`, whose `Quotes` field is exported precisely so that tests can
build one with a known slice:

```go
quoterepo.QuoteRepository{Quotes: []string{"foo"}}
```

Choosing which fields to export is how you design the API of a type.

## Embedding

A struct can embed another type by listing it without a field name:

```go
type config struct {
	baseconfig.Base

	LogPretty bool `envconfig:"LOG_PRETTY" default:"false"`
}
```

The fields and methods of `Base` are promoted: `config.Config.Host` reaches
`Base.Host` directly. This is composition with syntactic sugar, not inheritance.
There is no overriding, no `super`, and `config` is not a `Base`; it merely has
one. If you need to pass the embedded part somewhere, write
`config.Config.Base`.

## Interfaces

An interface is a set of method signatures. A type satisfies it by having those
methods. No declaration, no keyword:

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

`http.ServeMux` has a `ServeHTTP` method, so it is an `http.Handler`. The Twirp
server generated for this service also has one, so it is an `http.Handler` too,
and that is why `router.Handle(s.PathPrefix(), s)` compiles.

The generated code defines the service contract as an interface:

```go
type Quotes interface {
	GetQuote(context.Context, *GetQuoteRequest) (*GetQuoteResponse, error)
	GetQuoteFlaky(context.Context, *GetQuoteFlakyRequest) (*GetQuoteFlakyResponse, error)
	GetQuoteLate(context.Context, *GetQuoteLateRequest) (*GetQuoteLateResponse, error)
}
```

and `internal/handlers/handlers.go` satisfies it by having those three methods:

```go
type Quotes struct {
	ServerContext *appserver.ServerContext
}

func (h *Quotes) GetQuote(ctx context.Context, request *proto.GetQuoteRequest) (*proto.GetQuoteResponse, error) {
	return api.GetQuoteHandler(ctx, h.ServerContext, request)
}
```

Nothing in `handlers.go` mentions the interface. `main.go` passes
`&handlers.Quotes{...}` to `api.NewQuotesServer`, whose parameter type is the
interface, and the compiler checks that the methods match at that call. Add an
RPC to the proto, regenerate, and that call fails to compile until you add the
method. The compiler is your checklist.

### Small interfaces

Because satisfying an interface costs nothing, Go interfaces are tiny.
`error` has one method, `io.Reader` has one, `http.Handler` has one. Define
interfaces where they are consumed, not where types are implemented, and only
when you have a second implementation or a test that needs one.

This repository passes concrete pointers (`*QuoteRepository`,
`*QuoteService`) rather than interfaces, because there is one implementation of
each. The tests do not need mocks: they build a real repository with one quote.
When you do need to swap implementations, you introduce the interface at that
point, and existing types satisfy it retroactively.

### Compile-time assertion

To make satisfaction explicit and get an early error, add a line like:

```go
var _ api.Quotes = (*handlers.Quotes)(nil)
```

It declares a discarded variable of the interface type initialised with a typed
nil pointer. If the methods do not match, this line fails to compile. Useful in
the package that defines the implementation.

## Type assertions and switches

Going from an interface back to a concrete type:

```go
msg, ok := response["msg"].(string)    // ok is false if it is not a string

switch v := value.(type) {
case string:
	...
case int:
	...
default:
	...
}
```

The single-result form `response["msg"].(string)` panics on mismatch; prefer the
comma-ok form.

## Struct tags

```go
type QuoteRepository struct {
	Quotes []string `json:"quotes"`
}

LogPretty bool `envconfig:"LOG_PRETTY" default:"false"`
```

The backtick string after a field is a tag: metadata that libraries read with
reflection. `encoding/json` uses `json:"quotes"` to map the field to the JSON key
`quotes`; `envconfig` uses its tags to map to environment variables. Tags are the
Go answer to annotations and decorators for data mapping.

## Try it

1. Define `type QuoteSource interface { Gimme(context.Context) (string, error) }`
   in `quoteservice`, change `QuoteService.repo` to that type, and confirm
   `*quoterepo.QuoteRepository` still works with no change in `quoterepo`.
2. Write a second implementation, `fixedSource`, that always returns the same
   string, and use it in a `quoteservice` test.
3. Add the compile-time assertion for `handlers.Quotes` and then comment out one
   handler method. Read the error.
4. Add a `String() string` method to `NoQuotesError` and print it with `%v` and
   `%s`. Look up `fmt.Stringer`.

## Read next

- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Go Proverbs](https://go-proverbs.github.io): "The bigger the interface, the
  weaker the abstraction."
- [06. Context, goroutines and shutdown](06-context-goroutines-shutdown.md)
