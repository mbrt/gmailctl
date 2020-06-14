#!/bin/bash

set -e

# check that we are not committing with things still to generate
go generate ./...
# `generate` adds additional dependencies that we want to remove.
go mod tidy
git diff --exit-code