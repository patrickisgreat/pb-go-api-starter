module "alb" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/alb?ref=be137a59d4732673e40310a4883a92085d41bef4"

  name       = substr(local.name, 0, 32)
  vpc_id     = data.aws_vpc.this.id
  subnet_ids = var.internal ? data.aws_subnets.private.ids : data.aws_subnets.public.ids
  internal   = var.internal

  ingress_cidr_blocks = var.ingress_cidr_ranges
  certificate_arn     = var.certificate_arn

  target_port  = var.container_port
  health_check = { path = var.api_healthcheck_path }

  tags = local.tags
}
