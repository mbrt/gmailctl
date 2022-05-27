#!/bin/bash

docker build -t gmailctl-web -f web/Dockerfile .
docker run --rm -ti gmailctl-web cat web.js > web/web.js
