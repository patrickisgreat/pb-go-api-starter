variable "aws_profile" {
  description = "AWS CLI profile to deploy with"
  type        = string
}

variable "aws_region" {
  description = "AWS region to deploy to"
  type        = string
  default     = "eu-central-1"
}

variable "commit_sha" {
  description = "The commit being deployed, used for the docker image tag"
  type        = string
}

variable "environment" {
  description = "The environment to deploy to: development, staging or production"
  type        = string

  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "environment must be development, staging or production."
  }
}

variable "owner" {
  description = "Who owns the service. Written to resource tags and telemetry attributes."
  type        = string
  default     = "patrickisgreat"
}

# Networking
variable "vpc_name" {
  description = "Name tag of the VPC to deploy into"
  type        = string
  default     = "default"
}

variable "internal" {
  description = "Keep the load balancer internal to the VPC instead of internet-facing"
  type        = bool
  default     = false
}

variable "certificate_arn" {
  description = "ACM certificate for HTTPS on the load balancer. When null the service is served over plain HTTP."
  type        = string
  default     = null
}

variable "ingress_cidr_ranges" {
  description = "CIDR ranges allowed to reach the load balancer"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# ECS
variable "system" {
  description = "The name of the system"
  type        = string
  default     = "pb-go-api-starter"
}

variable "cpu" {
  description = "Task CPU limit in units"
  type        = number
  default     = 1024
}

variable "memory" {
  description = "Task memory limit in MB"
  type        = number
  default     = 2048
}

variable "api_healthcheck_path" {
  description = "The route the load balancer health checks"
  type        = string
  default     = "/-/health"
}

variable "ecs_min_tasks" {
  description = "Minimum number of api containers to run"
  type        = number
  default     = 1
}

variable "ecs_max_tasks" {
  description = "Maximum number of api containers to run"
  type        = number
  default     = 3
}

variable "scaleup_cooldown" {
  description = "Time to wait after scale up before scaling up again"
  type        = number
  default     = 60
}

variable "scaledown_cooldown" {
  description = "Time to wait after scale down before scaling down again"
  type        = number
  default     = 300
}

variable "container_port" {
  description = "Container port"
  type        = number
  default     = 8000
}

# Alerts
variable "alerts_enabled" {
  description = "Whether or not to create alarms for the environment"
  type        = bool
  default     = false
}

variable "alerts_runbook_url" {
  description = "The pb-go-api-starter runbook URL"
  type        = string
  default     = ""
}

variable "alert_emails" {
  description = "Email addresses that receive alarm notifications"
  type        = list(string)
  default     = []
}

variable "error_duration" {
  description = "Set the duration in minutes for the high errors alert"
  type        = number
  default     = 2
}

variable "error_threshold" {
  description = "Set the threshold percentage for the high errors alert"
  type        = number
  default     = 5
}

variable "latency_threshold" {
  description = "Set the p95 threshold in seconds for the high response time alert"
  type        = number
  default     = 3
}

variable "latency_duration" {
  description = "Set the duration in minutes for the high response time alert"
  type        = number
  default     = 2
}
