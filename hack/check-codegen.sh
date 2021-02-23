#!/bin/bash

set -e

# check that we are not committing with things still to generate
go get github.com/go-bindata/go-bindata/go-bindata
go generate ./...
git diff --exit-code -- pkg/ cmd/
