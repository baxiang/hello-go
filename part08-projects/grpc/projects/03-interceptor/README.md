# 项目03：Interceptor（拦截器）

演示 gRPC 拦截器的实际应用。

## 功能
1. **日志拦截器**：记录每个请求的耗时和结果
2. **认证拦截器**：验证请求中的 Token
3. **恢复拦截器**：捕获 panic 防止服务崩溃

## 运行
```bash
# 启动服务端
go run server/main.go

# 另开终端运行客户端
go run client/main.go
```

## 预期输出
客户端会演示三种场景：
1. 不带 Token → 认证失败
2. 错误 Token → 认证失败
3. 正确 Token → 请求成功

## 代码结构
```
03-interceptor/
├── proto/
│   └── interceptor.proto # 服务定义
├── server/
│   └── main.go           # 三个拦截器实现
├── client/
│   └── main.go           # 三种认证场景演示
└── go.mod
```

## 对应文档
- [05-拦截器与中间件](../docs/05-拦截器与中间件.md)