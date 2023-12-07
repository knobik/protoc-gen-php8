#!/usr/bin/env bash

# exit when any command fails
set -e

go build

LANGUAGE_TO_GEN=$1
INPUT_FILE=$2
PLUGIN_OPTIONS=""

if [[ $LANGUAGE_TO_GEN == "php8" ]]; then
  PLUGIN_OPTIONS="--experimental_allow_proto3_optional --plugin=protoc-gen-php8"
fi

BUILD_DIR="gen_${LANGUAGE_TO_GEN}"

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

bash -c "protoc ${PLUGIN_OPTIONS} --${LANGUAGE_TO_GEN}_out=${BUILD_DIR} ${INPUT_FILE}"