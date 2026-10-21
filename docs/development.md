# Development workflow

## Everyday commands

Every workflow is a [Task](https://taskfile.dev) task defined in `Taskfile.yml`.
`task --list-all` shows them all.

| Task                | Runs                                                        | Notes                                              |
| ------------------- | ----------------------------------------------------------- | -------------------------------------------------- |
| `task run`          | `air`                                                       | Hot reload. Rebuilds and restarts on save          |
| `task build`        | `task gen:rpc` then `go build -o cmd/app/bin/pb-go-api-starter` | Static binary, `GOOS` defaults to `darwin`     |
| `task gen:rpc`      | `protoc` with the Go and Twirp plugins                      | Wipes and recreates `internal/generated`; commit the result |
| `task gen:docs`     | `protoc-gen-doc` in Docker                                  | Rewrites `docs/pb_go_api_starter.md`               |
| `task format`       | `goimports -w -local <module>`                              | Groups imports: stdlib, third party, this module   |
| `task lint`         | `golangci-lint run`                                         | Runs `format` first                                |
| `task vet`          | `go vet ./...`                                              |                                                    |
| `task tidy`         | `go mod tidy`                                               |                                                    |
| `task test`         | `tidy`, `lint`, `vet`, then `go test`                       | Unit tests only                                    |
| `task test:e2e`     | `build`, then `go test ./end_to_end`                        | Spawns the binary                                  |
| `task upgrade`      | `go get -u ./...`                                           | Follow with `task tidy` and `task test`            |

Tasks declare `sources` and `generates`, so Task skips work whose inputs have not
changed. Delete `.task/` to force a full rerun.

## Hot reload

`.air.toml` tells `air` to watch `app`, `internal`, `pkg` and `rpc` for `.go`
changes, run `task build`, and restart `cmd/app/bin/pb-go-api-starter`. Test
files and `internal/generated` are excluded from the watch. Build errors go to
`build-errors.log`.

## Changing the API

1. Edit `rpc/pb_go_api_starter.proto`.
2. `task gen:rpc`.
3. Fix whatever the compiler now complains about in `internal/handlers`.
4. Implement, test, and run `task gen:docs`.

The [API guide](api.md#adding-an-rpc) has the detailed version.

## Code style

- `gofmt` formatting is not negotiable; `goimports` enforces it plus import
  grouping. Run `task format` before committing.
- Packages are named for what they provide (`quoterepo`, `quoteservice`), all
  lower case, no underscores.
- Handlers take `ctx` first and return `(response, error)`.
- Errors are returned, not logged and swallowed. Log at the point where the error
  is handled, with `logger.Get(ctx)` so that log lines carry trace context.
- Anything under `internal/` cannot be imported from outside this module. That is
  a Go rule, not a convention, and it is why almost everything lives there.

## Dependencies

`go.mod` pins direct dependencies; `go.sum` pins checksums for everything. After
adding an import, run `task tidy` so both files are updated. Every dependency is
a public module.

## Continuous integration and releases

`.github/workflows/ci.yml` runs on every pull request and on `main`. It is
three calls to reusable workflows in
[actions-toolkit](https://github.com/patrickisgreat/actions-toolkit), pinned to
`@v0`:

| Job             | Checks                                                                               |
| --------------- | ------------------------------------------------------------------------------------ |
| Go              | `gofmt`, `go vet`, `golangci-lint`, unit and end to end tests with `-race`, cross-compile for linux/amd64 and linux/arm64 |
| Terraform       | `terraform fmt`, `init` and `validate` of `terraform/`, `tflint`, `trivy` config scan |
| Container image | Builds the `Dockerfile` and scans the image with `trivy`; nothing is pushed         |

The Terraform validate job clones the tf-tools modules over SSH, so the
repository needs a `MODULES_SSH_KEY` secret holding a read-only deploy key for
`patrickisgreat/tf-tools`. Without it that one job fails and the others are
unaffected.

`.github/workflows/release.yml` runs on every push to `main`. release-please
reads the conventional commit messages since the last tag (`feat:`, `fix:` and
so on), opens or updates a release pull request carrying the version bump and
`CHANGELOG.md`, waits for its checks, and merges it. The merge creates the
`vX.Y.Z` tag and the GitHub release. Two settings make this work:

- Repository settings, Actions, General: allow GitHub Actions to create and
  approve pull requests.
- Optionally a `RELEASE_TOKEN` secret with a personal access token. Pushes made
  with the default token do not trigger other workflows, so without it the
  release commit itself is not CI-checked.

Commit messages therefore follow [Conventional Commits](https://www.conventionalcommits.org):
`feat(api): add ListQuotes` is a minor bump, `fix(quoterepo): ...` a patch, and
a `BREAKING CHANGE:` footer a major.

## Backstage

`catalog-info.yaml` registers the service and its API with Backstage, and
`mkdocs.yml` configures this documentation site for TechDocs. When you add a
page to `docs/`, add it to the `nav` in `mkdocs.yml`.
