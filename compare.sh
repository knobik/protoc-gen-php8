#!/usr/bin/env bash

# exit when any command fails
set -e

go build

rm -rf gen_php_old
mkdir -p gen_php_old

protoc --php_out=./gen_php_old protobuf/test.proto