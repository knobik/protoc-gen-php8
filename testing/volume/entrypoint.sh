#!/usr/bin/env bash

set -e

rm -rf gen
mkdir gen

PROTO_FILES=(
  "test.proto"
  "test_descriptors.proto"
  "test_import_descriptor_proto.proto"
  "test_reserved_enum_lower.proto"
  "test_reserved_enum_upper.proto"
  "test_reserved_enum_value_lower.proto"
  "test_reserved_enum_value_upper.proto"
  "test_reserved_message_lower.proto"
  "test_reserved_message_upper.proto"
  "test_service.proto"
  "test_service_namespace.proto"
  "test_wrapper_type_setters.proto"
)

for VALUE in "${PROTO_FILES[@]}"
do
  protoc --experimental_allow_proto3_optional --plugin=protoc-gen-php8 --php8_out=./gen "protobuf/$VALUE"
#  protoc --php_out=./gen "protobuf/$VALUE"
done

composer install

./vendor/bin/phpunit tests