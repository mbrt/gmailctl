#!/bin/bash

# Update the version number in cmd/gmailctl/cmd/version.go

set -e

VERSION=${1:?}

[[ "${VERSION}" =~  ^[0-9]+\.[0-9]+\.[0-9]+(-dev)?$ ]] || {
  echo "Invalid version format: ${VERSION}"
  exit 1
}

sed -i "s/version = \".*\"/version = \"${VERSION}\"/" cmd/gmailctl/cmd/version.go
