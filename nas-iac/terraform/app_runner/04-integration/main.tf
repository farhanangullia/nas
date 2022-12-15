################################################################################
# Integration Resources
################################################################################

data "aws_dynamodb_table" "allow_list_table" {
  name = var.allow_list_table_name
}

data "aws_dynamodb_table" "allow_list_requests_table" {
  name = var.allow_list_requests_table_name
}

data "aws_lambda_function" "ip_expiry_handler" {
  function_name = var.ip_expiry_handler_name
}

data "aws_lambda_function" "ip_retention_handler" {
  function_name = var.ip_retention_handler_name
}

resource "aws_lambda_event_source_mapping" "allow_list_requests_trigger" {
  event_source_arn       = data.aws_dynamodb_table.allow_list_requests_table.stream_arn
  function_name          = data.aws_lambda_function.ip_retention_handler.arn
  starting_position      = "LATEST"
  batch_size             = 1
  enabled                = true
  maximum_retry_attempts = 3
  filter_criteria {
    filter {
      pattern = "{ \"eventName\":[\"INSERT\"]}"
    }

    filter {
      pattern = "{ \"dynamodb\": { \"NewImage\": { \"Status\": { \"S\": [\"Failed\"] } } } }"
    }
  }
}

resource "aws_lambda_event_source_mapping" "allow_list_trigger" {
  event_source_arn       = data.aws_dynamodb_table.allow_list_table.stream_arn
  function_name          = data.aws_lambda_function.ip_expiry_handler.arn
  starting_position      = "LATEST"
  batch_size             = 1
  enabled                = true
  maximum_retry_attempts = 3
  filter_criteria {
    filter {
      pattern = "{ \"eventName\":[\"REMOVE\"]}"
    }
  }
}