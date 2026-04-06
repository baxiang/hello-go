# Wire 学习笔记

本目录包含完整的 Wire 依赖注入学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
wire/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Wire概述.md
│   ├── 02-基本概念.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-入门-HelloWorld/
│   ├── 02-进阶-Web应用/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Wire概述 → docs/02-基本概念 → docs/03-Provider与Injector

### 进阶路径
docs/04-高级特性 → docs/05-最佳实践 → docs/06-常见问题

### 实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-进阶-Web应用 → ...

## 本地开发环境

### 安装 Wire

```bash
# 安装 wire 工具
go install github.com/google/wire/cmd/wire@v0.6.0

# 验证安装
wire version
```

### 使用示例

```bash
# 在项目目录运行
cd projects/01-入门-HelloWorld
wire

# 生成的 wire_gen.go 文件会包含依赖注入代码
```

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Wire概述.md](./docs/01-Wire概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Wire | 0.6.0 |
| Go | 1.21+ |