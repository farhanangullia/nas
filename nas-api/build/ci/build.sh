#!/bin/bash

platform="${1:-amd64}"

# Build with Docker. Execute from root directory.
docker build -f deploy/docker/nas/Dockerfile -t nas --platform=$platform .

# Build AMD64
#docker build -f deploy/docker/nas/Dockerfile -t nas --platform=amd64 .

# Build ARM64
#docker build -f deploy/docker/nas/Dockerfile -t nas --platform=arm64 .