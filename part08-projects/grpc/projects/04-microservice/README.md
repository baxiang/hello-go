# 项目04：Microservice（微服务通信）

综合实践项目，演示生产环境常用特性。

## 功能
1. **TLS 加密**：服务端/客户端双向 TLS 认证
2. **健康检查**：内置 gRPC 健康检查服务
3. **服务统计**：记录请求数和连接数
4. **多实例支持**：通过 `-port` 和 `-id` 参数启动多个实例

## 运行
```bash
# 启动第一个服务实例
go run server/main.go -port 50054 -id server-1

# 另开终端启动第二个实例
go run server/main.go -port 50055 -id server-2

# 运行客户端
go run client/main.go
```

## 生成 TLS 证书（可选）
```bash
# 使用 openssl 生成自签名证书
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes -subj "/CN=localhost"
```

## 代码结构
```
04-microservice/
├── proto/
│   └── microservice.proto # 服务定义（含统计接口）
├── server/
│   └── main.go            # TLS + 健康检查 + 统计
├── client/
│   └── main.go            # TLS 连接 + 多请求演示
└── go.mod
```

## 对应文档
- [07-负载均衡](../docs/07-负载均衡.md)
- [08-安全与认证](../docs/08-安全与认证.md)
- [09-性能优化](../docs/09-性能优化.md)