data "aws_region" "current" {}
data "aws_partition" "current" {}

data "aws_ecr_repository" "this" {
  name = var.ecr_repository_name
}

data "aws_dynamodb_table" "allow_list_table" {
  name = var.allow_list_table_name
}

data "aws_dynamodb_table" "allow_list_requests_table" {
  name = var.allow_list_requests_table_name
}

################################################################################
# nas Config
################################################################################

resource "aws_ssm_parameter" "nas_ipset_config" {
  name        = "/nas/ipsets/config"
  description = "JSON config for target WAF IpSets"
  type        = "String"
  value       = file("${path.module}/nas_waf_ipset_config.json")
}

################################################################################
# AWS App Runner
################################################################################

# AppRunner Supporting Dependencies
resource "aws_apprunner_observability_configuration" "this" {
  observability_configuration_name = "nas-obs-config"

  trace_configuration {
    vendor = "AWSXRAY"
  }

}

################################################################################
# IAM Role - Access
################################################################################

data "aws_iam_policy_document" "access_assume_role" {
  statement {
    sid     = "AccessAssumeRole"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["build.apprunner.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "access" {
  name = "nas-app-runner-access-role"

  assume_role_policy    = data.aws_iam_policy_document.access_assume_role.json
  force_detach_policies = true

  tags = var.tags
}

data "aws_iam_policy_document" "access" {
  statement {
    sid = "ReadPrivateEcr"
    actions = [
      "ecr:BatchGetImage",
      "ecr:DescribeImages",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchCheckLayerAvailability"
    ]
    resources = [data.aws_ecr_repository.this.arn]

  }

  statement {

    sid = "AuthPrivateEcr"
    actions = [
      "ecr:DescribeImages",
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"]

  }
}

resource "aws_iam_policy" "access" {
  name = "app-runner-access-role-policy"

  policy = data.aws_iam_policy_document.access.json
}

resource "aws_iam_role_policy_attachment" "access" {
  policy_arn = aws_iam_policy.access.arn
  role       = aws_iam_role.access.name
}


################################################################################
# IAM Role - Instance
################################################################################

data "aws_iam_policy_document" "instance_assume_role" {
  statement {
    sid     = "InstanceAssumeRole"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["tasks.apprunner.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "instance" {
  name = "nas-app-runner-instance-role"

  assume_role_policy    = data.aws_iam_policy_document.instance_assume_role.json
  force_detach_policies = true

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "instance_xray" {
  policy_arn = "arn:${data.aws_partition.current.id}:iam::aws:policy/AWSXRayDaemonWriteAccess"
  role       = aws_iam_role.instance.name
}

data "aws_iam_policy_document" "instance_dynamodb" {
  statement {
    sid     = "AllAPIActionsOnAllowList"
    actions = ["dynamodb:*"]

    resources = [data.aws_dynamodb_table.allow_list_table.arn]
  }

  statement {
    sid     = "AllAPIActionsOnAllowListRequests"
    actions = ["dynamodb:*"]

    resources = [data.aws_dynamodb_table.allow_list_requests_table.arn]
  }
}

resource "aws_iam_policy" "instance_dynamodb" {
  name = "app-runner-instance-role-ddb-policy"

  policy = data.aws_iam_policy_document.instance_dynamodb.json
}

resource "aws_iam_role_policy_attachment" "instance_dynamodb" {
  policy_arn = aws_iam_policy.instance_dynamodb.arn
  role       = aws_iam_role.instance.name
}


resource "aws_apprunner_service" "this" {
  service_name = "ip-whitelisting-service"

  instance_configuration {
    instance_role_arn = aws_iam_role.instance.arn
  }

  observability_configuration {
    observability_configuration_arn = aws_apprunner_observability_configuration.this.arn
    observability_enabled           = true
  }

  network_configuration {
    ingress_configuration {
      is_publicly_accessible = false
    }
  }

  source_configuration {
    image_repository {
      image_configuration {
        port = var.image_port
      }
      image_identifier      = "${data.aws_ecr_repository.this.repository_url}:${var.image_tag}"
      image_repository_type = "ECR"
    }
    auto_deployments_enabled = var.auto_deployments_enabled

    authentication_configuration {
      access_role_arn = aws_iam_role.access.arn
    }

  }

  health_check_configuration {
    interval = var.health_check_interval
    timeout  = var.health_check_timeout
    path     = var.health_check_path
    protocol = var.health_check_protocol
  }

  depends_on = [
    aws_iam_role.access,
    aws_iam_role.instance,
    aws_iam_role_policy_attachment.access,
    aws_iam_role_policy_attachment.instance_xray,
    aws_iam_role_policy_attachment.instance_dynamodb
  ]
}
################################################################################
# VPC Ingress Configuration
################################################################################

resource "aws_apprunner_vpc_ingress_connection" "this" {
  name        = "ip-whitelisting-service-ingress"
  service_arn = aws_apprunner_service.this.arn

  ingress_vpc_configuration {
    vpc_id          = var.ingress_vpc_id
    vpc_endpoint_id = var.ingress_vpc_endpoint_id
  }

  tags = var.tags
}