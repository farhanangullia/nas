variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

variable "trusted_lambda_role_arns" {
  description = "List of lambda role arns to allow assume role for updating ipset"
  type        = list(string)
}