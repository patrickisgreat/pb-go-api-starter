# Configuration

All configuration comes from environment variables, read once at start-up by
`config.Load` in `internal/config/config.go`. Code reads the resulting
`config.Config` struct; nothing calls `os.Getenv` directly.

## How values reach the process

**Locally** Task loads env files before running any task:

```yaml
dotenv: ['.env/.env.{{.ENV}}', '.env/.env.local']
```

`ENV` selects the file. `task run` sets `ENV=local`, the test tasks set
`ENV=test`, and you can set it yourself, for example
`ENV=development task test:e2e:ci`. `.env/.env.local` is always loaded last and
wins on conflicts, so it is the place for personal overrides.

**In AWS** Terraform injects the variables into the ECS task definition. Log
level and trace sampling are chosen per environment in `terraform/locals.tf`.

## Service variables

| Variable            | Set in                | Meaning                                                              |
| ------------------- | --------------------- | -------------------------------------------------------------------- |
| `APP_NAME`          | env files, Terraform  | Service name for logs and telemetry resource attributes              |
| `HOST`              | env files             | Interface to bind. `localhost` for local and test                    |
| `PORT`              | env files, Terraform  | Port to bind. `8000` everywhere                                      |
| `ENVIRONMENT`       | env files, Terraform  | `local`, `test`, `development`, `staging` or `production`            |
| `LOG_LEVEL`         | env files, Terraform  | `debug` (dev), `info` (staging), `error` (production and test)       |
| `LOG_PRETTY`        | `.env.local`          | `true` prints human readable logs. Default `false` means JSON        |
| `OTEL_SDK_DISABLED` | `.env.local`, `.env.test` | `true` switches off telemetry export so local runs need no collector |

`APP_NAME`, `HOST`, `PORT`, `ENVIRONMENT` and `LOG_LEVEL` are fields of
`baseconfig.Base` in `internal/baseconfig`, the set every service needs.
`LOG_PRETTY` is specific to this service and is the pattern to copy when adding
a new setting:

```go
type config struct {
    baseconfig.Base

    LogPretty bool `envconfig:"LOG_PRETTY" default:"false"`
}
```

## End to end test variables

| Variable   | Meaning                                                                 |
| ---------- | ----------------------------------------------------------------------- |
| `API_HOST` | Host the end to end tests target. When unset they start a local binary  |
| `API_PORT` | Port the tests target. Default `8000`                                   |

`.env/.env.development`, `.env.staging` and `.env.production` contain only these
two variables, pointing at the deployed instance for that environment.

## Telemetry variables

In deployed environments Terraform sets the standard OpenTelemetry variables
(`OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_TRACES_EXPORTER`, `OTEL_TRACES_SAMPLER_ARG`
and so on) plus `OTEL_RESOURCE_ATTRIBUTES`, which identifies the environment,
owner, repository and commit using the OpenTelemetry semantic convention names.
The full list is `telemetry_config` in `terraform/locals.tf`. Trace sampling
ratios are 100% in development, 10% in staging and 0.1% in production.

`internal/telemetry` reads these at start-up. Setting `OTEL_TRACES_EXPORTER`,
`OTEL_METRICS_EXPORTER` or `OTEL_LOGS_EXPORTER` to `none` switches off that one
signal; `OTEL_SDK_DISABLED=true` switches off all three.

## Adding a setting

1. Add a field to the `config` struct with an `envconfig` tag and a `default`.
2. Add it to `.env/.env.local` and `.env/.env.test` if local runs need a
   non-default value.
3. If production needs it, add it to the container environment in
   `terraform/main.tf`.
4. Document it in the table above.
