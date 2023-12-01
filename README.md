# Build

```shell
go build .
```

# Usage

```shell
protoc --plugin=protoc-gen-php8 --php8_out=./gen your-proto-file.proto
```

# TODO:

1. [x] oneof field
2. [x] proto 3.15 optional field option (https://stackoverflow.com/questions/42622015/how-to-define-an-optional-field-in-protobuf-3)
3. [ ] gRPC services
4. [x] ~~Metadata~~