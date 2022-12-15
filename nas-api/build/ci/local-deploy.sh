#!/bin/bash

# Run Docker container. Image needs to be available locally.
docker run --env-file .env -p 8080:8080 -it nas