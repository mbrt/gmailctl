#!/bin/bash

# check that we are not committing with things still to generate
go generate ./...
git diff --exit-code
