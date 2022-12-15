## Deployment steps

1. Deploy networking (tf)
2. Deploy dependencies (tf)
3. Deploy Lambda function (sls)
4. Deploy iam role and waf ipset in target account(s) (tf)
   1. Deploy this in every AWS account where the WAF IpSet can be updated
5. Deploy app runner / ecs (tf)
6. Deploy integration (tf)