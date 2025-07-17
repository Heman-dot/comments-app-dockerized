#!/usr/bin/env bash

set -eExuo pipefail

# This basic script is needed to cope with the PIE's limitation regarding
# docker and each student's home directory. One cannot mount with docker
# files from home directory.

cp prometheus.yml /tmp/prometheus.yml
docker compose up -d || docker-compose up -d
