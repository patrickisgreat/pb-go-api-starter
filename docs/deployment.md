# Deployment

The service ships as a container image and runs on AWS ECS. Terraform in
`terraform/` describes the service, its scaling and its alerts, with one
configuration directory per environment. The reusable modules come from
[tf-tools](https://github.com/patrickisgreat/tf-tools); this repository holds
only the composition.

## Container image

`Dockerfile` is a two stage build:

1. **build** on `golang:alpine`. Downloads modules, then compiles a static binary
   with `CGO_ENABLED=0` and `-ldflags="-w -s"` to strip debug information.
2. **final** on `scratch`. Copies in only the CA certificate bundle and the
   binary. There is no shell and no package manager in the image.

Build and run it locally:

```sh
docker build --platform linux/amd64 -t pb-go-api-starter:local .

docker run --rm -p 8000:8000 \
  -e APP_NAME=pb-go-api-starter -e HOST=0.0.0.0 -e PORT=8000 \
  -e ENVIRONMENT=local -e LOG_LEVEL=debug -e OTEL_SDK_DISABLED=true \
  pb-go-api-starter:local
```

`.dockerignore` keeps docs, terraform, env files and the task cache out of the
build context.

## Image registry

Terraform creates the ECR repository itself (module `ecr`) and prints its URL
as the `ecr_repository_url` output. Tags are immutable, so tag with the full
commit SHA: every deployment is then traceable to a commit and rollbacks are a
matter of applying with an older `commit_sha`.

On a brand new environment the repository has to exist before the service can
start, so the first apply is done in two steps:

```sh
terraform apply -target=module.ecr -var-file=config/development/terraform.tfvars -var commit_sha=<sha>
docker tag pb-go-api-starter:local $(terraform output -raw ecr_repository_url):<sha>
docker push $(terraform output -raw ecr_repository_url):<sha>
```

then the full apply below.

## Terraform layout

```
terraform/
  main.tf           ECR, IAM roles, cluster, service and autoscaling; log group, service security group
  load_balancer.tf  the application load balancer, target group and listeners
  alerts.tf         SNS topic and the CloudWatch alarms on the load balancer metrics
  variables.tf      system name, networking, TLS certificate, sizing, scaling, alert thresholds
  locals.tf         per environment log level, trace sampling, telemetry env vars, image URI
  data.tf           VPC and subnet lookups by tag
  outputs.tf        load balancer hostname, log group, ECR repository URL
  terraform.tf      backend and provider configuration
  config/
    development/    config.remote (state backend) and terraform.tfvars
    staging/
    production/
```

Every module is a tf-tools module referenced by git source over SSH, for example
`git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-service?ref=main`:
`ecr`, `iam-role`, `ecs-cluster`, `ecs-service`, `ecs-autoscaling`, `alb`,
`alb-alarms` and `sns`. The only raw resources left here are the log group and
the service's security group.
tf-tools is a private repository, so `terraform init` needs an SSH key that can
read it: your own GitHub key locally, and in CI a read-only deploy key stored as
the `MODULES_SSH_KEY` secret. Pin `ref` to a tag or commit before relying on it
in production, and bump every module's `ref` together.

Environment specific values are resolved two ways. `config/<env>/terraform.tfvars`
supplies the AWS profile and environment name. `locals.tf` then maps the
environment to log level and trace ratio.

The VPC is found by its `Name` tag (`vpc_name`, default `default`) and subnets by
a `Tier` tag of `Private` or `Public`. Tasks always run in private subnets; the
load balancer uses public subnets unless `internal = true`. Set
`certificate_arn` to an ACM certificate to serve HTTPS; port 80 then redirects
to 443.

## Applying

```sh
cd terraform
terraform init -backend-config=config/development/config.remote
terraform plan  -var-file=config/development/terraform.tfvars -var commit_sha=<sha> -out plan.tfplan
terraform apply plan.tfplan
```

Replace `development` with `staging` or `production`. The state backend for each
environment is an S3 bucket with a DynamoDB lock table, named in `config.remote`.
Create those once per environment before the first `init`.

`terraform output api_host` prints the load balancer hostname; put it in
`.env/.env.<environment>` as `API_HOST` so that the end to end tests can target
the deployment.

## Alerts

`alerts.tf` wires the tf-tools `alb-alarms` module to an `sns` topic: alarms
on the percentage of requests answered with a 5xx, p95 response time, and
unhealthy targets, with `alert_emails` listing the subscribed addresses. The
thresholds and evaluation windows are variables with defaults in `variables.tf`,
and `alerts_runbook_url` is included in every alarm description so that the
notification links straight to the runbook. `GetQuoteFlaky` and `GetQuoteLate`
exist so these alarms can be triggered on purpose in a non-production
environment.

## Verifying a deployment

After applying, run the end to end suite against the environment:

```sh
ENV=development task test:e2e:ci
```

This uses `API_HOST` and `API_PORT` from `.env/.env.development`. Then confirm the
service is logging to the CloudWatch log group printed by `terraform output
log_group`, and that traces are arriving at whichever backend the OTLP collector
forwards to.
