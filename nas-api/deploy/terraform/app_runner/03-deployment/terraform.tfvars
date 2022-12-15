ecr_repository_name = "nas"
allow_list_table_name = "IpAllowList"
allow_list_requests_table_name = "IpAllowListRequests"

image_port = "8080"
image_tag  = "v1.0.0"

health_check_interval = 15
health_check_timeout  = 5
health_check_path     = "/nas/api/v2/healthz"
health_check_protocol = "HTTP"

ingress_vpc_id          = "REPLACE_ME"
ingress_vpc_endpoint_id = "REPLACE_ME"

