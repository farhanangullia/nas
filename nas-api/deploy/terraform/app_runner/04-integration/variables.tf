variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

variable "ip_expiry_handler_name" {
  description = "Name of lambda function"
  type        = string
}

variable "ip_retention_handler_name" {
  description = "Name of lambda function"
  type        = string
}

variable "allow_list_table_name" {
  description = "Name of DynamoDB Table"
  type        = string
}

variable "allow_list_requests_table_name" {
  description = "Name of DynamoDB table"
  type        = string
}