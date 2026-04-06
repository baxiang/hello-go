# gRPC 概述

gRPC 是 Google 开源的高性能 RPC 框架，基于 HTTP/2 和 Protocol Buffers。

## 核心特性

- **高性能**：基于 HTTP/2，支持流式传输
- **强类型**：使用 Protocol Buffers 定义接口
- **多语言支持**：支持 Go、Java、Python、C++ 等
- **双向流**：支持客户端流、服务端流、双向流

## 架构

```
Client                  Server
  │                       │
  │ ─── Request ────────> │
  │ <─── Response ─────── │
```

## Protocol Buffers

定义服务接口：

```protobuf
syntax = "proto3";

package helloworld;

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

message HelloRequest {
  string name = 1;
}

message HelloReply {
  string message = 1;
}
```

## 快速开始

### 安装

```bash
# 安装 protoc
brew install protobuf

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 编译 proto

```bash
protoc --go_out=. --go-grpc_out=. helloworld.proto
```

### Server

```go
package main

import (
    "context"
    "net"
    "google.golang.org/grpc"
    pb "path/to/helloworld"
)

type server struct {
    pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
    return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
    lis, _ := net.Listen("tcp", ":50051")
    s := grpc.NewServer()
    pb.RegisterGreeterServer(s, &server{})
    s.Serve(lis)
}
```

### Client

```go
package main

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "path/to/helloworld"
)

func main() {
    conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    defer conn.Close()
    
    client := pb.NewGreeterClient(conn)
    resp, _ := client.SayHello(context.Background(), &pb.HelloRequest{Name: "World"})
    println(resp.Message)
}
```

## 版本信息

| 组件 | 版本 |
|------|------|
| gRPC-Go | 1.62.0 |
| Protobuf | 3.25.3 |

在下一章中，我们将深入学习 Protocol Buffers。
