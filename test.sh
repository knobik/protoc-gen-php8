#!/usr/bin/env bash

set -e

go get -d -v
GOOS=linux GOARCH=amd64 go build -o testing/volume/protoc-gen-php8

if [[ "$(docker images -q protoc-gen-php8-tests 2> /dev/null)" == "" ]]; then
  docker build -t protoc-gen-php8-tests testing/.
fi
docker run -v "${PWD}/testing/volume:/tests" protoc-gen-php8-tests