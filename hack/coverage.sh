#!/bin/bash

set -e

COV=$(mktemp --suffix=.cov)
HTML=$(mktemp --suffix=.html)

go test -v -coverpkg=./... -coverprofile="${COV}" ./...
go tool cover -func="${COV}"
go tool cover -html="${COV}" -o "${HTML}"

echo "Find results in ${HTML}"
