# gRPC 学习笔记

本目录包含完整的 gRPC 学习资料，包括理论文档和配套实战项目。每个项目都是完整可运行的代码，可直接编译执行。

## 目录结构

```
grpc/
├── docs/                    # 理论文档（按学习顺序）
│   ├── README.md           # 系列导航
│   ├── 01-gRPC概述.md
│   ├── 02-Protocol-Buffers.md
│   ├── 03-Go服务端与客户端开发.md
│   ├── 04-流式RPC.md
│   ├── 05-拦截器与中间件.md
│   ├── 06-错误处理.md
│   ├── 07-负载均衡.md
│   ├── 08-安全与认证.md
│   └── 09-性能优化.md
└── projects/               # 实战项目（配套文档）
    ├── 01-helloworld/      # 一元调用（对应 01、03）
    ├── 02-stream-rpc/      # 流式传输（对应 04）
    ├── 03-interceptor/     # 拦截器（对应 05）
    └── 04-microservice/    # 微服务通信（对应 07、08、09）
```

## 学习路径

| 阶段 | 文档 | 实战项目 | 预计时间 |
|------|------|----------|----------|
| 基础 | 01-gRPC概述 → 02-Protocol-Buffers | - | 1.5 小时 |
| 入门 | 03-Go服务端与客户端开发 | [01-helloworld](./projects/01-helloworld/) | 1 小时 |
| 进阶 | 04-流式RPC | [02-stream-rpc](./projects/02-stream-rpc/) | 1 小时 |
| 进阶 | 05-拦截器与中间件 | [03-interceptor](./projects/03-interceptor/) | 1 小时 |
| 进阶 | 06-错误处理 | - | 0.5 小时 |
| 生产 | 07-负载均衡 → 08-安全与认证 → 09-性能优化 | [04-microservice](./projects/04-microservice/) | 2 小时 |

## 环境准备

```bash
# 安装 protoc
brew install protobuf

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 验证
protoc --version
```

## 快速开始

以 HelloWorld 项目为例：

```bash
cd projects/01-helloworld

# 生成 protobuf 代码（已预生成，如修改 proto 需重新执行）
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/helloworld.proto

# 启动服务端
go run server/main.go

# 另开终端运行客户端
go run client/main.go
```

## 版本信息

| 组件 | 版本 |
|------|------|
| gRPC-Go | 1.81.1 |
| Protobuf | 3.25.3 |
| Go | 1.21+ |

## 参考

- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers](https://protobuf.dev/)
- [gRPC-Go GitHub](https://github.com/grpc/grpc-go)