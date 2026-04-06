# gRPC 学习笔记

本目录包含完整的 gRPC 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
grpc/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-gRPC概述.md
│   ├── 02-Protocol-Buffers.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-入门-HelloWorld/
│   ├── 02-进阶-流式RPC/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-gRPC概述 → docs/02-Protocol-Buffers → docs/03-Go客户端开发

### 进阶路径
docs/04-流式RPC → docs/05-拦截器与中间件 → docs/06-错误处理

### 实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-进阶-流式RPC → ...

## 本地开发环境

### 安装 Protobuf 编译器

```bash
# macOS
brew install protobuf

# 或下载二进制文件
# https://github.com/protocolbuffers/protobuf/releases

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 运行示例

```bash
# 编译 proto 文件
protoc --go_out=. --go-grpc_out=. proto/helloworld.proto

# 运行服务端
go run server/main.go

# 运行客户端
go run client/main.go
```

### 连接地址

- gRPC 服务端口: localhost:50051

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-gRPC概述.md](./docs/01-gRPC概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| gRPC-Go | 1.62.0 |
| Protobuf | 3.25.3 |
| Go | 1.21+ |