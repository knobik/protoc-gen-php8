# Build

```shell
go build .
```

# Usage

```shell
protoc --plugin=protoc-gen-php8 --php8_out=./gen your-proto-file.proto
```

# TODO:

- [x] ~~oneof field~~
- [x] ~~proto 3.15 optional field option (https://stackoverflow.com/questions/42622015/how-to-define-an-optional-field-in-protobuf-3)~~
- [x] ~~metadata~~
- [x] ~~service interface support~~
- [x] ~~wrapped fields~~
- [x] ~~deprecation comment~~
- [ ] well known types
- [ ] copy comments from proto files to generated classes
- [ ] prefix reserved keywords
