# etcd 学习笔记

本目录包含完整的 etcd 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
etcd/
├── docs/                    # 理论文档（9篇系统性学习指南）
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-etcd概述.md
│   ├── 02-核心概念-数据模型.md
│   ├── 03-核心概念-Raft共识.md
│   ├── 04-核心概念-集群管理.md
│   ├── 05-Go-客户端基础.md
│   ├── 06-Go-客户端-KV操作.md
│   ├── 07-Go-客户端-Watch和Lease.md
│   ├── 08-Go-客户端-事务和锁.md
│   └── 09-生产部署.md
├── projects/               # 实战项目（4个渐进式项目）
│   ├── 01-入门-HelloWorld/
│   ├── 02-入门-配置中心/
│   ├── 03-进阶-服务发现/
│   └── 04-高级-分布式锁/
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-etcd概述 → docs/02-核心概念-数据模型 → docs/03-核心概念-Raft共识

### 进阶路径
docs/04-核心概念-集群管理 → docs/05-Go-客户端基础 → docs/06-Go-客户端-KV操作

### 高级路径
docs/07-Go-客户端-Watch和Lease → docs/08-Go-客户端-事务和锁 → docs/09-生产部署

### 开发实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-入门-配置中心 → ...

## 本地开发环境

### 启动 etcd 服务

```bash
# 使用 Docker（推荐）
docker run -d --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.17 \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster s1=http://0.0.0.0:2380 \
  --initial-cluster-token tkn \
  --initial-cluster-state new

# 或使用 Homebrew (macOS)
brew install etcd
etcd

# 或下载二进制文件
# https://github.com/etcd-io/etcd/releases
```

### 连接地址

- etcd 客户端端口: http://localhost:2379
- etcd 对等端口: http://localhost:2380

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-etcd概述.md](./docs/01-etcd概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| etcd | 3.5.17 |
| etcd/client/v3 | 3.5.17 |
| Go | 1.21+ |