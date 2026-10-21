output "api_host" {
  description = "Hostname of the deployed API; use it as API_HOST for the end to end tests"
  value       = module.alb.dns_name
}

output "log_group" {
  description = "CloudWatch log group for the service"
  value       = aws_cloudwatch_log_group.this.name
}

output "ecr_repository_url" {
  description = "Push the image here, tagged with the commit SHA"
  value       = module.ecr.repository_url
}
