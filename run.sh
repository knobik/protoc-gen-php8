#!/usr/bin/env bash

# exit when any command fails
set -e

go build

rm -rf gen
mkdir -p gen

protoc --plugin=protoc-gen-php8 --php8_out=./gen protobuf/test.proto