# Getting started

This page gets the service running locally and confirms it answers requests.

## 1. Install prerequisites

- [Go](https://go.dev/doc/install) 1.26 or newer. Check with `go version`.
- [Task](https://taskfile.dev), the command runner used for every workflow here:

  ```sh
  go install github.com/go-task/task/v3/cmd/task@latest
  ```

  Make sure `$(go env GOPATH)/bin` is on your `PATH` so that `task` and the
  `protoc` plugins installed below can be found.

- Docker, only for building the container image or regenerating the API docs.

## 2. Install the toolchain

```sh
task deps:macos
```

On macOS this installs `protobuf` and `golangci-lint` with Homebrew, then
`air`, `goimports`, `protoc-gen-go` and `protoc-gen-twirp` with `go install`.
On Linux install `protoc` and `golangci-lint` from your package manager and run
`task deps:proto` for the Go tools.

## 3. Generate code and download dependencies

```sh
task gen:rpc    # protoc -> internal/generated/pb_go_api_starter/{api,session}
task install    # go get -t ./...
```

The generated code is committed, so a fresh clone compiles as is; `task gen:rpc`
rewrites it whenever a `.proto` file changes, and the build task runs it for
you. All dependencies are public modules, so no credentials are needed.

## 4. Run it

```sh
task run
```

This runs `air`, which builds the binary, starts it, and rebuilds on every save.
Configuration comes from `.env/.env.local`: the server listens on
`localhost:8000` with pretty debug logs and telemetry export disabled.

## 5. Make a request

```sh
curl -i http://localhost:8000/-/health

curl -s -X POST -H "Content-Type: application/json" \
  http://localhost:8000/twirp/proto.patrickisgreat.pb_go_api_starter.api.Quotes/GetQuote \
  --data '{}'
```

You should see `OK` from the health check and a JSON object with a `quote` field
from the RPC.

## 6. Run the tests

```sh
task test        # unit tests, preceded by tidy, lint and vet
task test:e2e    # builds the binary, starts it, runs end_to_end/
```

## Common problems

| Symptom                                              | Fix                                                        |
| ---------------------------------------------------- | ---------------------------------------------------------- |
| `internal/generated` does not match the proto files  | Run `task gen:rpc` and commit the result.                  |
| `protoc-gen-go: program not found`                   | Add `$(go env GOPATH)/bin` to `PATH`, or rerun `task deps:proto`. |
| `task: command not found`                            | Same `PATH` issue as above.                                |
| Port 8000 already in use                             | Change `PORT` in `.env/.env.local`.                        |
