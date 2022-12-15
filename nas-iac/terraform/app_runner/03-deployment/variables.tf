variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

variable "ecr_repository_name" {
  description = "Name of the ECR repository for App Runner"
  type        = string
}

variable "image_port" {
  description = "Port to expose for the container"
  type        = string
}

variable "image_tag" {
  description = "Tag of app runner image"
  type        = string
}

variable "auto_deployments_enabled" {
  description = "Set to True for automated deployments when new image is released"
  type        = bool
  default     = false
}

variable "health_check_interval" {
  description = "Time interval, in seconds, between health checks."
  type        = number
}

variable "health_check_timeout" {
  description = "Time, in seconds, to wait for a health check response before deciding it failed."
  type        = number
}

variable "health_check_path" {
  description = "URL to send requests to for health checks."
  type        = string
}

variable "health_check_protocol" {
  description = "IP protocol that App Runner uses to perform health checks for your service."
  type        = string
}

variable "ingress_vpc_id" {
  description = "The ID of the VPC that is used for the VPC endpoint."
  type        = string
}

variable "ingress_vpc_endpoint_id" {
  description = "The ID of the VPC endpoint that your App Runner service connects to."
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