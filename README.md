# Build

```shell
go build .
```

# Usage

```shell
protoc --plugin=protoc-gen-php8 --php8_out=./gen your-proto-file.proto
```

# TODO:

1. [ ] oneof field
2. [x] ~~Metadata~~