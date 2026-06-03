# 入门 - HelloWorld

本项目演示 etcd 的基本操作，包括连接、读写、删除等基础功能。

## 项目概述

- 学习如何使用 Go 客户端连接 etcd
- 掌握基本的 Put、Get、Delete 操作
- 理解 etcd 的数据模型

## 目录结构

```
01-getting-started-helloworld/
├── README.md           # 项目说明
├── main.go            # 主程序
└── go.mod             # Go 模块定义
```

## 运行步骤

### 1. 启动 etcd

```bash
docker run -d --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.17 \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379
```

### 2. 运行程序

```bash
cd 01-getting-started-helloworld
go run main.go
```

## 预期输出

```
Connected to etcd
Put succeeded: revision 2
Get result: Hello, etcd!
Deleted: 1 key(s)
Key no longer exists
```

## 学习要点

1. **客户端连接**：使用 `clientv3.New()` 创建客户端
2. **超时控制**：使用 `context.WithTimeout()` 控制操作超时
3. **Put 操作**：`cli.Put()` 写入键值
4. **Get 操作**：`cli.Get()` 读取键值
5. **Delete 操作**：`cli.Delete()` 删除键值

## 相关文档

- [docs/01-etcd概述.md](../../docs/01-etcd概述.md)
- [docs/05-Go-客户端基础.md](../../docs/05-Go-客户端基础.md)