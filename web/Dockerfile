FROM golang:1.17-alpine

RUN apk --no-cache add git && \
    go install github.com/gopherjs/gopherjs@v1.17.2

WORKDIR /src/project
COPY . .

RUN gopherjs build -m ./web
