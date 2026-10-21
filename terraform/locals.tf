locals {
  environment = var.environment
  short_env_names = {
    production  = "prod"
    staging     = "stage"
    development = "dev"
  }
  short_env_name = local.short_env_names[local.environment]

  tags = {
    Application = var.system
    GitRepo     = "github.com/patrickisgreat/pb-go-api-starter"
    ManagedBy   = "Terraform"
    Owner       = var.owner
    Environment = local.environment
  }

  log_levels = {
    production  = "error"
    staging     = "info"
    development = "debug"
  }
  log_level = local.log_levels[local.environment]

  trace_ratios = {
    production  = "0.001"
    staging     = "0.1"
    development = "1"
  }
  trace_ratio = local.trace_ratios[local.environment]

  # Resource attributes identify the service in whatever backend the OTLP
  # collector forwards to. They use the OpenTelemetry semantic convention
  # names so that no backend specific mapping is needed.
  resource_attributes = {
    "deployment.environment.name" = local.environment
    "service.namespace"           = var.owner
    "service.version"             = var.commit_sha
    "vcs.repository.url.full"     = "https://${local.tags.GitRepo}"
  }

  telemetry_config = {
    OTEL_EXPORTER_OTLP_PROTOCOL = "grpc"
    OTEL_LOGS_EXPORTER          = "otlp"
    OTEL_METRICS_EXPORTER       = "otlp"
    OTEL_RESOURCE_ATTRIBUTES    = join(",", [for k, v in local.resource_attributes : "${k}=${v}"])
    OTEL_SERVICE_NAME           = var.system
    OTEL_TRACES_EXPORTER        = "otlp"
    OTEL_TRACES_SAMPLER         = "parentbased_traceidratio"
    OTEL_TRACES_SAMPLER_ARG     = local.trace_ratio
  }

  name            = "${var.system}-${local.environment}"
  container_image = "${module.ecr.repository_url}:${var.commit_sha}"
  ecs_cluster     = "${var.system}-${local.short_env_name}"
}
