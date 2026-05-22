# Redis 学习笔记

本目录包含完整的 Redis 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
redis/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Redis概述.md
│   ├── 02-数据类型与操作.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-入门-HelloWorld/
│   ├── 02-进阶-缓存系统/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Redis概述 → docs/02-数据类型与操作 → docs/03-Go客户端基础

### 进阶路径
docs/04-持久化机制 → docs/05-主从复制 → docs/06-哨兵与集群

### 实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-进阶-缓存系统 → ...

## 本地开发环境

### 启动 Redis

```bash
# 使用 Docker
docker run -d --name redis \
  -p 6379:6379 \
  redis:7.2.4

# 或使用 Homebrew (macOS)
brew install redis
brew services start redis

# 或使用配置文件启动
docker run -d --name redis \
  -p 6379:6379 \
  -v /path/to/redis.conf:/usr/local/etc/redis/redis.conf \
  redis:7.2.4 \
  redis-server /usr/local/etc/redis/redis.conf
```

### 连接地址

- Redis 端口: localhost:6379

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Redis概述.md](./docs/01-Redis概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Redis | 7.2.4 |
| go-redis | 9.5.1 |
| Go | 1.21+ |