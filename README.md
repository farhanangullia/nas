# NAS

This monorepo is a collection of components and services which makes up the network access service.

`nas-api` is a HTTP serving application written in Go for interacting with clients. It conforms to the clean architecture approach and leverages on Go kit framework.

`nas-consumers` contains event driven serverless functions written in Python for executing business logic and serverless framework templates for deployments.

`nas-iac` contains infrastructure as code written in Terraform to deploy the services.

## CI

GitLab CI is used to configure CI/CD for this project. The root CI has the base logic for triggering child pipelines according to the sub directory that changed.