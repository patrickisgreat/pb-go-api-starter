module "alerts_topic" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/sns?ref=be137a59d4732673e40310a4883a92085d41bef4"
  count  = var.alerts_enabled ? 1 : 0

  name         = "${local.name}-alerts"
  display_name = "${var.system} ${local.environment} alerts"
  subscriptions = [
    for email in var.alert_emails : { protocol = "email", endpoint = email }
  ]
  tags = merge(local.tags, { Component = "alerts" })
}

module "alarms" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/alb-alarms?ref=be137a59d4732673e40310a4883a92085d41bef4"
  count  = var.alerts_enabled ? 1 : 0

  name                     = local.name
  load_balancer_arn_suffix = module.alb.arn_suffix
  target_group_arn_suffix  = module.alb.target_group_arn_suffix
  alarm_actions            = [module.alerts_topic[0].arn]
  runbook_url              = var.alerts_runbook_url == "" ? null : var.alerts_runbook_url

  error_rate_threshold_percent  = var.error_threshold
  error_rate_evaluation_periods = var.error_duration
  latency_threshold_seconds     = var.latency_threshold
  latency_evaluation_periods    = var.latency_duration

  tags = merge(local.tags, { Component = "alerts" })
}
