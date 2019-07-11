#!/bin/bash

FILES="*.go"

echo
echo "monitoring:"
echo $FILES
echo
echo

while true; do
    inotifywait -q -e modify $FILES
    echo
    echo
    echo
    echo
    go test -run Observer01
done
