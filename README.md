# Build

```shell
go build .
```

# Usage

```shell
protoc --plugin=protoc-gen-php8 --php8_out=./gen your-proto-file.proto
```

### Options
```shell
--php8_opt=MessageParentClass="App\\Protobuf\\BaseMessage"
```

# TODO:

- [x] ~~variant types~~
- [x] ~~oneof field~~
- [x] ~~proto 3.15 optional field option (https://stackoverflow.com/questions/42622015/how-to-define-an-optional-field-in-protobuf-3)~~
- [x] ~~metadata~~
- [x] ~~service interface support~~
- [x] ~~wrapped fields~~
- [x] ~~deprecation comments~~
- [x] ~~well known types~~
- [x] ~~fix repeated enum~~
- [ ] prefix reserved keywords
- [ ] copy comments from proto files to generated classes
- [ ] extensions
- [ ] refactor class properties into separate templates
