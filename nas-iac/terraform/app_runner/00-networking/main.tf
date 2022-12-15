################################################################################
# Supporting Resources
################################################################################

data "aws_availability_zones" "available" {}

locals {
  name     = "nas"
  vpc_cidr = "10.0.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 3.0"

  name = local.name
  cidr = local.vpc_cidr

  azs             = local.azs
  private_subnets = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 4, k)]
  public_subnets  = [for k, v in local.azs : cidrsubnet(local.vpc_cidr, 8, k + 48)]

  enable_nat_gateway      = false
  single_nat_gateway      = true
  enable_dns_hostnames    = true
  map_public_ip_on_launch = true

  tags = var.tags
}

module "vpc_endpoints" {
  source  = "terraform-aws-modules/vpc/aws//modules/vpc-endpoints"
  version = "~> 3.18.1"

  vpc_id             = module.vpc.vpc_id
  security_group_ids = [module.vpc_endpoints_security_group.security_group_id]

  endpoints = {
    apprunner = {
      service = "apprunner.requests"
      # private_dns_enabled = true
      subnet_ids = module.vpc.private_subnets
      tags       = var.tags
    },
  }

  tags = var.tags
}

# module "security_group" {
#   source  = "terraform-aws-modules/security-group/aws"
#   version = "~> 4.0"

#   name        = local.name
#   description = "Security group for AppRunner connector"
#   vpc_id      = module.vpc.vpc_id

#   ingress_rules = ["all-all"]
#   egress_rules       = ["all-all"]
#   egress_cidr_blocks = module.vpc.private_subnets_cidr_blocks

#   tags = var.tags
# }

module "vpc_endpoints_security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.16.2"

  name        = "${local.name}-vpc-endpoints"
  description = "Security group for VPC Endpoints"
  vpc_id      = module.vpc.vpc_id

  ingress_rules      = ["https-443-tcp"]
  ingress_cidr_blocks = ["0.0.0.0/0"]
  egress_rules       = ["all-all"]
  egress_cidr_blocks = [module.vpc.vpc_cidr_block]

  tags = var.tags
}