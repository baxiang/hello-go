# Protocol Buffers

Protocol Buffers（protobuf）是 Google 的结构化数据序列化机制。

## Proto3 语法

```protobuf
syntax = "proto3";
package myapp;

message Person {
  string name = 1;
  int32 age = 2;
  repeated string emails = 3;
}
```

## 数据类型映射

| Proto类型 | Go类型 |
|-----------|--------|
| string | string |
| int32 | int32 |
| int64 | int64 |
| bool | bool |
| bytes | []byte |

## 编译安装

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 编译命令

```bash
protoc --go_out=. --go-grpc_out=. hello.proto
```

参考：https://protobuf.dev/
