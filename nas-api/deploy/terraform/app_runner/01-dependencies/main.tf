data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

################################################################################
# Supporting Resources
################################################################################

resource "aws_ecr_repository" "this" {
  name                 = "nas"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = var.tags
}


resource "aws_dynamodb_table" "allow_list_requests" {
  name     = "IpAllowListRequests"
  hash_key = "Id"

  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  read_capacity  = 10
  write_capacity = 10

  attribute {
    name = "Id"
    type = "S"
  }

  tags = var.tags
}

resource "aws_dynamodb_table" "allow_list" {
  name      = "IpAllowList"
  hash_key  = "Ip"
  range_key = "AwsAccountId"


  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  read_capacity  = 10
  write_capacity = 10

  attribute {
    name = "Ip"
    type = "S"
  }

  attribute {
    name = "AwsAccountId"
    type = "S"
  }

  ttl {
    attribute_name = "Expiry"
    enabled        = true
  }

  tags = var.tags
}

 
# Common

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    sid     = "LambdaAssumeRole"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy" "ssm_read_only" {
  name = "AmazonSSMReadOnlyAccess"
}

data "aws_iam_policy" "lambda_dynamodb" {
  name = "AWSLambdaInvocation-DynamoDB"
}

# IP Retention handler

resource "aws_iam_role" "lambda_ip_retention_handler" {
  name = "nas-lambda-ip-retention-handler-role"

  assume_role_policy    = data.aws_iam_policy_document.lambda_assume_role.json
  force_detach_policies = true

  tags = var.tags
}

data "aws_iam_policy_document" "lambda_ip_retention_handler" {

  statement {

    sid = "BasicExecutionLogGroup"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = ["arn:aws:logs:*:${data.aws_caller_identity.current.account_id}:*"]
  }

  statement {

    sid = "BasicExecutionLogStream"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/nas-ip-retention-handler*:*"]
  }

  statement {

    sid = "DynamoDB"
    actions = [
      "dynamodb:*"
    ]

    resources = [aws_dynamodb_table.allow_list_requests.arn, aws_dynamodb_table.allow_list.arn]
  }

  statement {
    sid = "AssumeRole"
    actions = [
      "sts:AssumeRole"
    ]
    resources = ["arn:aws:iam::*:role/nas-waf-ipset-role"]

  }
}

resource "aws_iam_policy" "lambda_ip_retention_handler" {
  name = "nas-ip-retention-handler-policy"

  policy = data.aws_iam_policy_document.lambda_ip_retention_handler.json
}

resource "aws_iam_role_policy_attachment" "lambda_ip_retention_handler" {
  policy_arn = aws_iam_policy.lambda_ip_retention_handler.arn
  role       = aws_iam_role.lambda_ip_retention_handler.name
}

resource "aws_iam_role_policy_attachment" "retention_ssm_read_only" {
  policy_arn = data.aws_iam_policy.ssm_read_only.arn
  role       = aws_iam_role.lambda_ip_retention_handler.name
}

resource "aws_iam_role_policy_attachment" "retention_lambda_dynamodb" {
  policy_arn = data.aws_iam_policy.lambda_dynamodb.arn
  role       = aws_iam_role.lambda_ip_retention_handler.name
}

# Expiry handler


resource "aws_iam_role" "lambda_ip_expiry_handler" {
  name = "nas-lambda-ip-expiry-handler-role"

  assume_role_policy    = data.aws_iam_policy_document.lambda_assume_role.json
  force_detach_policies = true

  tags = var.tags
}

data "aws_iam_policy_document" "lambda_ip_expiry_handler" {

  statement {

    sid = "BasicExecutionLogGroup"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = ["arn:aws:logs:*:${data.aws_caller_identity.current.account_id}:*"]
  }

  statement {

    sid = "BasicExecutionLogStream"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/nas-ip-expiry-handler*:*"]
  }

  statement {

    sid = "DynamoDB"
    actions = [
      "dynamodb:*"
    ]

    resources = [aws_dynamodb_table.allow_list.arn]
  }

  statement {
    sid = "AssumeRole"
    actions = [
      "sts:AssumeRole"
    ]
    resources = ["arn:aws:iam::*:role/nas-waf-ipset-role"]

  }
}

resource "aws_iam_policy" "lambda_ip_expiry_handler" {
  name = "nas-ip-expiry-handler-policy"

  policy = data.aws_iam_policy_document.lambda_ip_expiry_handler.json
}

resource "aws_iam_role_policy_attachment" "lambda_ip_expiry_handler" {
  policy_arn = aws_iam_policy.lambda_ip_expiry_handler.arn
  role       = aws_iam_role.lambda_ip_expiry_handler.name
}

resource "aws_iam_role_policy_attachment" "expiry_ssm_read_only" {
  policy_arn = data.aws_iam_policy.ssm_read_only.arn
  role       = aws_iam_role.lambda_ip_expiry_handler.name
}

resource "aws_iam_role_policy_attachment" "expiry_lambda_dynamodb" {
  policy_arn = data.aws_iam_policy.lambda_dynamodb.arn
  role       = aws_iam_role.lambda_ip_expiry_handler.name
}

