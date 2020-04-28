#!/bin/bash

set -xe
cd $(dirname $(readlink -f $0))

SRC=munkres.go
PATCHES=(
    's@"github.com/cpmech/gosl/chk"@@'
    's@"github.com/cpmech/gosl/io"@@'
    's@"github.com/cpmech/gosl/utl"@@'
    's/chk\.//g'
    's/io\.//g'
    's/utl\.//g'
)

# Download the latest sources.
wget -O "${SRC}" https://github.com/cpmech/gosl/raw/master/graph/munkres.go

# Apply patches.
for p in ${PATCHES[@]}; do
    sed -i ${p} "${SRC}"
done
gofmt -w "${SRC}"
