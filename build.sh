#!/usr/bin/env bash

set -ex

docker buildx build --platform linux/amd64 --tag gcr.io/gadget-core-production/s2-decoder:latest --progress=plain .