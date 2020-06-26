#!/bin/bash

set -e

# check that we are not committing with things still to generate
go generate ./...
git diff --exit-code -- pkg/ cmd/
