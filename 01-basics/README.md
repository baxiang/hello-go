# Part 01: Go语言基础

本部分是Go语言学习的起点，涵盖从环境搭建到工程实践的完整内容。

## 学习目标

- 独立搭建 Go 开发环境并熟练使用常用工具
- 掌握 Go 语法基础与流程控制
- 理解数组、切片、Map、字符串、结构体等核心数据类型
- 编写函数、使用指针、理解内存模型
- 掌握接口、错误处理、包管理、泛型
- 能够开发命令行工具和简单项目

## 课程目录

| 章节 | 内容 | 所属目录 | 时间 |
|------|------|----------|------|
| 1.1 | 环境搭建与工具链 | [01-environment-tools](./01-environment-tools/) | 1-2天 |
| 1.2 | 常用工具 | [01-environment-tools](./01-environment-tools/) | 1-2天 |
| 1.3 | 语法基础 | [02-basic-syntax](./02-basic-syntax/) | 2-3天 |
| 1.4 | 流程控制 | [02-basic-syntax](./02-basic-syntax/) | 1-2天 |
| 1.5 | 数组与切片 | [03-data-structures](./03-data-structures/) | 2天 |
| 1.6 | Map | [03-data-structures](./03-data-structures/) | 1-2天 |
| 1.7 | 字符串与字节处理 | [03-data-structures](./03-data-structures/) | 2天 |
| 1.8 | 结构体与方法 | [03-data-structures](./03-data-structures/) | 2天 |
| 1.9 | 函数 | [04-functions-pointers](./04-functions-pointers/) | 2天 |
| 1.10 | 指针 | [04-functions-pointers](./04-functions-pointers/) | 1-2天 |
| 1.11 | 接口 | [05-interfaces-errors](./05-interfaces-errors/) | 2天 |
| 1.12 | 错误处理 | [05-interfaces-errors](./05-interfaces-errors/) | 1-2天 |
| 1.13 | 包管理 | [06-engineering](./06-engineering/) | 1-2天 |
| 1.14 | 时间处理 | [06-engineering](./06-engineering/) | 1-2天 |
| 1.15 | 泛型（独立模块） | [07-generics](./07-generics/) | 1周 |
| 1.16 | Cobra 命令行工具 | [06-engineering](./06-engineering/) | 2天 |

## 推荐学习路径

### 第1周：环境与语法

```
01-environment-tools/
├── 1.1 环境搭建与工具链
│   ├── Go 安装与版本管理（goenv / mise / asdf）
│   ├── GOPROXY 配置、GOPATH 与 Go Modules
│   └── VS Code / GoLand 开发环境
└── 1.2 常用工具
    ├── gofmt / goimports 格式化
    ├── go vet / staticcheck / golangci-lint
    └── Delve 调试器、pprof 性能分析

02-basic-syntax/
├── 1.3 语法基础
│   ├── 变量与常量（var、const、:=、iota）
│   ├── 基本数据类型与运算符
│   └── 格式化输入输出
└── 1.4 流程控制
    ├── if / switch 条件判断
    ├── for 循环与 range 遍历
    └── defer 延迟执行
```

### 第2周：数据类型与函数

```
03-data-structures/
├── 1.5 数组与切片
├── 1.6 Map
├── 1.7 字符串与字节处理
│   ├── strings / bytes 包常用函数
│   ├── strconv 类型转换
│   └── 字符串拼接性能优化
└── 1.8 结构体与方法
    ├── 结构体定义、匿名字段嵌入
    ├── 值接收者 vs 指针接收者
    └── struct tag 与 JSON 序列化

04-functions-pointers/
├── 1.9 函数
│   ├── 多返回值、可变参数
│   ├── 匿名函数与闭包
│   └── init 函数与 defer 顺序
└── 1.10 指针
    ├── 声明与解引用
    ├── new vs make
    └── 值传递 vs 指针传递
```

### 第3周：接口、工程与实战

```
05-interfaces-errors/
├── 1.11 接口
│   ├── 隐式实现与鸭子类型
│   ├── 类型断言与类型开关
│   └── 接口嵌套与设计原则
└── 1.12 错误处理
    ├── error 接口、error wrapping（%w）
    ├── errors.Is / As / Join
    └── panic / recover

06-engineering/
├── 1.13 包管理
│   ├── go.mod / go.sum 详解
│   ├── 依赖版本管理
│   └── GOPRIVATE 私有仓库
├── 1.14 时间处理
│   ├── time.Time 格式化（2006-01-02 15:04:05）
│   ├── 时区处理与时间运算
│   └── Timer / Ticker 定时器
└── 1.16 Cobra 命令行工具
    ├── 命令与子命令
    ├── Flags 参数解析
    └── Todo CLI 实战
```

### 第4周：泛型系统

```
07-generics/
├── 2.1 泛型入门            — 类型参数、泛型函数与类型
├── 2.2 类型约束深入        — ~、comparable、约束组合
├── 2.3 泛型数据结构        — Stack/Queue/Set/LinkedList/LRU
├── 2.4 泛型算法            — Filter/Map/Reduce/GroupBy
├── 2.5 标准库泛型          — slices/maps/cmp 深入
└── 2.6 实战与最佳实践      — 何时用/何时不用/性能陷阱
```

## 学习检查清单

### 环境与工具
- [ ] Go 环境正确安装并配置
- [ ] VS Code / GoLand 开发环境就绪
- [ ] 会使用 gofmt、go vet、golangci-lint
- [ ] 掌握 Delve 调试基础

### 基础语法
- [ ] 掌握变量声明（var、:=）、常量定义（const、iota）
- [ ] 熟练使用基本数据类型与运算符
- [ ] 掌握 if/switch 条件判断、for 循环与 range
- [ ] 理解 defer 执行时机

### 数据类型
- [ ] 理解数组和切片的区别
- [ ] 掌握切片截取、append、copy 操作
- [ ] 理解切片底层原理与扩容机制
- [ ] 掌握 Map 的创建、CRUD 与遍历
- [ ] 熟练使用 strings 包与 bytes 包
- [ ] 掌握 strconv 类型转换
- [ ] 掌握结构体定义与实例化
- [ ] 理解值接收者与指针接收者的区别
- [ ] 掌握 struct tag 与 JSON 序列化

### 函数与指针
- [ ] 能够定义各种形式的函数
- [ ] 掌握多返回值和可变参数
- [ ] 理解闭包的概念与陷阱
- [ ] 掌握指针的声明与使用
- [ ] 理解值传递与指针传递的区别
- [ ] 区分 new 和 make

### 接口与错误
- [ ] 理解接口的隐式实现机制
- [ ] 掌握类型断言和类型开关
- [ ] 理解接口嵌套组合
- [ ] 掌握 error 接口与自定义错误
- [ ] 理解错误包装、errors.Is/As/Join
- [ ] 理解 panic/recover 的使用场景

### 工程实践
- [ ] 理解 Go Modules 工作原理
- [ ] 掌握依赖版本管理与私有仓库配置
- [ ] 熟练使用 time 包进行时间处理
- [ ] 能够编写泛型函数和泛型类型
- [ ] 理解类型约束的设计
- [ ] 使用 slices、cmp 等泛型标准库
- [ ] 能够使用 Cobra 构建命令行工具

## 入门项目推荐

### 1. 计算器 CLI
**难度**: ⭐
**技能点**: 基础语法、函数、错误处理
**描述**: 实现支持加减乘除的命令行计算器

### 2. 单词计数器
**难度**: ⭐⭐
**技能点**: 文件读取、Map、字符串处理
**描述**: 统计文本文件中单词出现频率

### 3. 待办事项管理
**难度**: ⭐⭐
**技能点**: 切片、Map、CLI 开发、文件操作
**描述**: 使用 JSON 文件存储的待办事项管理工具

### 4. 简单缓存系统
**难度**: ⭐⭐
**技能点**: Map、时间处理、指针
**描述**: 实现带过期时间的 in-memory 缓存

### 5. 链表库
**难度**: ⭐⭐⭐
**技能点**: 结构体、指针、接口、泛型
**描述**: 实现泛型单向/双向链表

## 下一步学习

完成本部分后，建议继续学习：

- **[02-advanced](../02-advanced/)** — Goroutine 并发、Channel、Context、标准库进阶

---

**预计学习时间**: 约 5-6 周（每天 2-3 小时）
