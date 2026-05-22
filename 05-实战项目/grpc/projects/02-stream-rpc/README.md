# 项目02：Stream RPC

演示 gRPC 的三种流式传输模式。

## 功能
1. **服务端流**：服务端向客户端推送多条消息
2. **客户端流**：客户端向服务端发送多条消息
3. **双向流**：双方同时发送流式数据

## 运行
```bash
# 启动服务端
go run server/main.go

# 另开终端运行客户端
go run client/main.go
```

## 代码结构
```
02-stream-rpc/
├── proto/
│   └── stream.proto      # 流式服务定义
├── server/
│   └── main.go           # 三种流式方法实现
├── client/
│   └── main.go           # 三种流式调用示例
└── go.mod
```

## 对应文档
- [04-流式RPC](../docs/04-流式RPC.md)