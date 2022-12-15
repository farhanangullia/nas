
################################################################################
# WAF IPSet
################################################################################

resource "aws_wafv2_ip_set" "nas_ipv4_ipset" {
  name               = "nas-ipv4"
  description        = "IPv4 ipset for nas"
  scope              = "CLOUDFRONT"
  ip_address_version = "IPV4"
  addresses          = [""]

  lifecycle {
    ignore_changes = [
      addresses,
    ]
  }

  tags = var.tags
}


################################################################################
# IAM Role - Access
################################################################################

data "aws_iam_policy_document" "waf_ipset_assume_role" {
  statement {
    sid     = "AccessAssumeRole"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = var.trusted_lambda_role_arns
    }

  }
}

resource "aws_iam_role" "waf_ipset" {
  name = "nas-waf-ipset-role"

  assume_role_policy    = data.aws_iam_policy_document.waf_ipset_assume_role.json
  force_detach_policies = true

  tags = var.tags
}

data "aws_iam_policy_document" "waf_ipset" {
  statement {
    sid = "AllowUseOfAWSWAF"
    actions = [
      "wafv2:Get*",
      "wafv2:UpdateIpSet"
    ]
    resources = [aws_wafv2_ip_set.nas_ipv4_ipset.arn]

  }
}

resource "aws_iam_policy" "waf_ipset" {
  name = "nas-waf-ipset-role-policy"

  policy = data.aws_iam_policy_document.waf_ipset.json
}

resource "aws_iam_role_policy_attachment" "waf_ipset" {
  policy_arn = aws_iam_policy.waf_ipset.arn
  role       = aws_iam_role.waf_ipset.name
}