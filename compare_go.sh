#!/usr/bin/env bash

# exit when any command fails
set -e

go build

rm -rf gen_go
mkdir -p gen_go

protoc --experimental_allow_proto3_optional --go_out=./gen_go protobuf/test.proto