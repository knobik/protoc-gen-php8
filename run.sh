#!/usr/bin/env bash

# exit when any command fails
set -e

go build

rm -rf gen_php_8
mkdir -p gen_php_8

protoc --experimental_allow_proto3_optional --plugin=protoc-gen-php8 --php8_out=./gen_php_8 protobuf/tests/test.proto