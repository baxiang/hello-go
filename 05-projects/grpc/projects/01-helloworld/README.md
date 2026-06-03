# 项目01：HelloWorld

最基础的 gRPC 服务，演示 Unary RPC（一元调用）。

## 功能
- 服务端提供 `SayHello` 接口
- 客户端发送名称，服务端返回问候消息

## 运行
```bash
# 启动服务端
go run server/main.go

# 另开终端运行客户端
go run client/main.go
```

## 代码结构
```
01-helloworld/
├── proto/
│   └── helloworld.proto  # Protobuf 定义
├── server/
│   └── main.go           # 服务端实现
├── client/
│   └── main.go           # 客户端实现
└── go.mod
```

## 对应文档
- [01-gRPC概述](../docs/01-gRPC概述.md)
- [03-Go服务端与客户端开发](../docs/03-Go服务端与客户端开发.md)