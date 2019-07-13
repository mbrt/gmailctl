#!/bin/bash

set -e

# check that we are not committing with things still to generate
go generate -mod vendor ./...
git diff --exit-code
