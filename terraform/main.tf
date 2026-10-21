# Reusable building blocks come from github.com/patrickisgreat/tf-tools. Every
# module source is pinned to the same tf-tools commit; bump them together when
# upgrading, and switch to a release tag once tf-tools publishes one that
# contains these modules. What remains here is the composition specific to
# this service.

module "ecr" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/ecr?ref=be137a59d4732673e40310a4883a92085d41bef4"

  name                       = var.system
  image_tag_mutability       = "IMMUTABLE"
  keep_last_n_images         = 30
  expire_untagged_after_days = 14
  tags                       = local.tags
}

module "execution_role" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/iam-role?ref=be137a59d4732673e40310a4883a92085d41bef4"

  name                  = "${local.name}-execution"
  description           = "Lets ECS pull the ${var.system} image and write its logs"
  trusted_role_services = ["ecs-tasks.amazonaws.com"]
  managed_policy_arns   = ["arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"]
  tags                  = local.tags
}

data "aws_iam_policy_document" "execute_command" {
  statement {
    actions = [
      "ssmmessages:CreateControlChannel",
      "ssmmessages:CreateDataChannel",
      "ssmmessages:OpenControlChannel",
      "ssmmessages:OpenDataChannel",
    ]
    resources = ["*"]
  }
}

module "task_role" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/iam-role?ref=be137a59d4732673e40310a4883a92085d41bef4"

  name                  = "${local.name}-task"
  description           = "Permissions of the running ${var.system} container"
  trusted_role_services = ["ecs-tasks.amazonaws.com"]
  inline_policies = {
    execute-command = data.aws_iam_policy_document.execute_command.json
  }
  tags = local.tags
}

module "cluster" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-cluster?ref=be137a59d4732673e40310a4883a92085d41bef4"

  name = local.ecs_cluster
  tags = local.tags
}

resource "aws_cloudwatch_log_group" "this" {
  name              = "/ecs/${local.name}"
  retention_in_days = 30
  tags              = local.tags
}

resource "aws_security_group" "service" {
  name        = "${local.name}-service"
  description = "ECS tasks for ${local.name}"
  vpc_id      = data.aws_vpc.this.id
  tags        = local.tags

  ingress {
    description     = "From load balancer"
    from_port       = var.container_port
    to_port         = var.container_port
    protocol        = "tcp"
    security_groups = [module.alb.security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

module "api_service" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-service?ref=be137a59d4732673e40310a4883a92085d41bef4"

  name        = var.system
  cluster_arn = module.cluster.arn
  launch_type = "FARGATE"
  cpu         = tostring(var.cpu)
  memory      = tostring(var.memory)

  # Autoscaling owns the running count from here on, so start at its floor.
  desired_count          = var.ecs_min_tasks
  execution_role_arn     = module.execution_role.arn
  task_role_arn          = module.task_role.arn
  enable_execute_command = true

  subnet_ids         = data.aws_subnets.private.ids
  security_group_ids = [aws_security_group.service.id]
  target_group_arn   = module.alb.target_group_arn
  lb_container_name  = var.system
  lb_container_port  = var.container_port

  container_definitions = jsonencode([
    {
      name      = var.system
      image     = local.container_image
      essential = true
      portMappings = [
        { containerPort = var.container_port, protocol = "tcp" }
      ]
      environment = [
        for k, v in merge(local.telemetry_config, {
          APP_NAME    = var.system
          LOG_LEVEL   = local.log_level
          ENVIRONMENT = local.environment
          PORT        = var.container_port
        }) : { name = k, value = tostring(v) }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.this.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = var.system
        }
      }
    }
  ])

  tags = merge(local.tags, { Component = "api" })
}

module "scaling" {
  source = "git::ssh://git@github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-autoscaling?ref=be137a59d4732673e40310a4883a92085d41bef4"

  cluster_name = module.cluster.name
  service_name = module.api_service.service_name

  min_capacity       = var.ecs_min_tasks
  max_capacity       = var.ecs_max_tasks
  cpu_target_percent = 40
  scale_out_cooldown = var.scaleup_cooldown
  scale_in_cooldown  = var.scaledown_cooldown
  tags               = local.tags
}
