#!/bin/bash

# This script is used to move a package from one directory to another.

function move() {
    local name=$1
    local src=internal/${name}
    local dst=internal/engine/${name}

    mv "${src}" "${dst}"
    for file in $(rg -l "${src}"); do
        sed -i "s@${src}@${dst}@g" "${file}"
    done
}

move api
move apply
move cfgtest
move config
move export
move filter
move gmail
move graph
move label
move parser
move rimport
