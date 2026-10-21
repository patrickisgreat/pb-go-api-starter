# 07. HTTP, protobuf and Twirp

The service speaks HTTP through `net/http`, defines its contract in protobuf, and
uses Twirp to generate the glue between them. This chapter shows what each layer
does, what the generated code is, and where the line falls between code you
write and code you never touch.

## `net/http` in three types

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

A `Handler` gets a `*Request` and writes to a `ResponseWriter`. A `ServeMux`
routes paths to handlers and is itself a `Handler`. A `Server` binds an address
and dispatches connections to a `Handler`. That is the entire model.

`internal/appserver/router.go`:

```go
func NewRouter(s api.TwirpServer) http.Handler {
	router := http.NewServeMux()
	router.Handle(s.PathPrefix(), s)
	return addHealthCheckRoute(router)
}

func addHealthCheckRoute(router *http.ServeMux) *http.ServeMux {
	router.Handle("/-/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "OK")
		if err != nil {
			http.Error(w, "health check failed", 500)
		}
	}))
	return router
}
```

Two things worth pausing on:

- `http.HandlerFunc` is a function type with a `ServeHTTP` method that calls the
  function. It adapts a plain function into a `Handler`. This "function type
  that satisfies an interface" trick appears throughout the standard library.
- Middleware is just a function from `Handler` to `Handler`. In `main.go`,
  `telemetry.InstrumentHandler("Quotes", router)` wraps the router so that
  every request is traced and measured before reaching it. You can chain as
  many as you like.

`ServeMux` patterns since Go 1.22 can include methods and path parameters
(`"GET /quotes/{id}"`), which makes it adequate for most REST APIs without a
third party router.

## Protobuf: the contract

`rpc/pb_go_api_starter.proto` defines messages and a service:

```proto
message GetQuoteRequest {
  proto.patrickisgreat.common.session.UserSession user_session = 1;
}

message GetQuoteResponse {
    string quote = 1;
}

service Quotes {
    rpc GetQuote (GetQuoteRequest) returns (GetQuoteResponse) {}
    ...
}
```

Protobuf is a language-neutral schema. The numbers are field tags used in the
binary encoding; never reuse or renumber them once clients exist. Fields have
zero values (`""`, `0`, `false`, `nil` for messages), and there is no way to tell
"not set" from "set to zero" for scalars, which is why `delay_ms` uses
`google.protobuf.Int32Value`, a wrapper message that can be `nil`.

`protoc` compiles this file. Two plugins produce Go:

- **`protoc-gen-go`** writes `pb_go_api_starter.pb.go`: a struct per message
  with exported fields (`Quote string`), getters (`GetQuote()`) that are safe on
  a nil receiver, and the marshalling code.
- **`protoc-gen-twirp`** writes `pb_go_api_starter.twirp.go`: the `Quotes`
  interface, `NewQuotesServer`, `NewQuotesJSONClient`,
  `NewQuotesProtobufClient` and the routing.

`task gen:rpc` runs it:

```sh
protoc pb_go_api_starter.proto --proto_path=rpc --go_out=./internal/generated --twirp_out=./internal/generated
```

The output lands in `internal/generated/pb_go_api_starter/api` because of
`option go_package="pb_go_api_starter/api";` in the proto. It is git ignored
and regenerated on every build, so treat it as read-only.

### Getters versus fields

```go
request.GetUserSession()            // nil-safe: returns nil if request is nil
request.UserSession                 // panics if request is nil
request.GetDelayMs().GetValue()     // safe even when delay_ms was not sent
```

Prefer getters when reading, especially through nested optional messages. Use
fields when constructing:

```go
&proto.GetQuoteResponse{Quote: quote}
```

## Twirp: the glue

Twirp maps each RPC to `POST /twirp/<package>.<Service>/<Method>`, accepting
`application/json` or `application/protobuf`. The generated server:

1. Parses the path to find the method.
2. Decodes the body into the request struct.
3. Calls your implementation with a `context.Context` carrying the request.
4. Encodes the response, or converts a returned error to Twirp error JSON with
   the matching HTTP status.

All you provide is a value satisfying the `Quotes` interface, which is
`handlers.Quotes`. The interceptor option in `main.go`,
`telemetry.WithTwirpServerInterceptor()`, wraps every method call in a
trace span.

Twirp also generates clients. Another Go service would call this one with:

```go
client := api.NewQuotesJSONClient("http://localhost:8000", &http.Client{})
resp, err := client.GetQuote(ctx, &api.GetQuoteRequest{})
```

and get typed responses and typed errors with no HTTP code in sight.

## The layers, again

```
net/http           connections, routing, middleware       (standard library)
Twirp generated    path → method, decode, encode, errors   (never edited)
handlers.Quotes    satisfies the interface, delegates      (one line per method)
api                validates, orchestrates                  (yours)
quoteservice       business rules, error translation       (yours)
quoterepo          data                                     (yours)
```

When something goes wrong, this table tells you where to look. A 404 with
`bad_route` is the Twirp layer: wrong path, wrong verb or wrong content type.
A 400 `invalid_argument` is `api`. A 500 `internal` is `quoteservice` or a
panic.

## Try it

1. Run the server and call `GetQuote` with `-X GET`, then with
   `Content-Type: text/plain`, then with a misspelled method. Read each error.
2. Send a protobuf request: `echo -n '' | curl -s -X POST -H "Content-Type: application/protobuf" --data-binary @- http://localhost:8000/twirp/proto.patrickisgreat.pb_go_api_starter.api.Quotes/GetQuote | xxd`.
   An empty body is a valid empty message, and the response comes back in
   binary protobuf.
3. Write a tiny `cmd/client/main.go` that uses `NewQuotesJSONClient` to call
   the running server and prints the quote.
4. Add a logging middleware `func(next http.Handler) http.Handler` that prints
   the method and path, and wrap the router with it in `main.go`.
5. Open `internal/generated/pb_go_api_starter/api/pb_go_api_starter.twirp.go`
   and find where `GetQuote` is routed. It is long but not complicated.

## Read next

- [net/http documentation](https://pkg.go.dev/net/http), the package overview
- [Twirp: Go implementation guide](https://twitchtv.github.io/twirp/docs/intro.html)
- [Protocol Buffers: Go generated code](https://protobuf.dev/reference/go/go-generated/)
- [08. Standard library tour](08-standard-library-tour.md)
