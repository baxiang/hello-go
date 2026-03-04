# Go 语言学习大纲 (2026)

## 第一部分：Go 语言基础

### 1.1 环境搭建与工具链
- Go 安装与版本管理 (goenv, asdf)
- GOPATH 与 Go Modules 理解
- 常用开发工具 (VS Code + Go 插件，GoLand)
- go 命令详解 (run, build, test, fmt, vet)
- 工作区设置与项目结构

### 1.2 语法基础
- 变量与常量声明
- 基本数据类型 (int, float, string, bool)
- 运算符与表达式
- 格式化输入输出
- 注释规范

### 1.3 流程控制
- if/else 条件语句
- switch 语句 (包括 type switch)
- for 循环 (三种形式)
- break, continue, goto
- defer 语句

### 1.4 数组与切片
- 数组定义与使用
- 切片的创建与操作
- 切片的扩容机制
- copy 和 append 函数
- 多维数组与切片

### 1.5 Map
- Map 的创建与初始化
- 增删改查操作
- Map 的遍历
- Map 与切片的组合使用
- sync.Map 简介

### 1.6 函数
- 函数定义与调用
- 多返回值
- 可变参数
- 匿名函数
- 闭包与作用域

### 1.7 指针
- 指针基础概念
- new 和 make 的区别
- 指针与函数参数传递
- 指针的安全性

---

## 第二部分：核心特性

### 2.1 结构体与方法
- 结构体定义与实例化
- 匿名字段与结构体嵌入
- 结构体标签 (struct tags)
- 方法定义 (值接收者 vs 指针接收者)
- Stringer 接口实现

### 2.2 接口
- 接口定义与实现
- 空接口 interface{}
- 类型断言与类型开关
- 接口嵌套
- 接口设计原则

### 2.3 错误处理
- error 接口
- 自定义错误类型
- errors 包 (Is, As, Join)
- panic 与 recover
- 错误处理最佳实践

### 2.4 包管理
- Go Modules 详解
- go.mod 与 go.sum 解析
- 依赖版本管理
- 私有仓库配置
- 发布自己的包

### 2.5 泛型 (Go 1.18+)
- 泛型语法基础
- 类型参数与类型约束
- 泛型函数与泛型类型
- 标准库中的泛型
- 泛型使用场景与限制

---

## 第三部分：并发编程

### 3.1 Goroutine
- Goroutine 概念与原理
- Goroutine 调度模型 (GMP 模型)
- Goroutine 与线程的区别
- Goroutine 泄露问题

### 3.2 Channel
- Channel 的创建与使用
- 无缓冲 vs 有缓冲 Channel
- Channel 的方向
- Channel 的关闭
- range 遍历 Channel
- select 多路复用

### 3.3 同步原语
- sync.WaitGroup
- sync.Mutex 与 sync.RWMutex
- sync.Cond
- sync.Once
- atomic 原子操作

### 3.4 Context
- Context 接口详解
- Context 的创建与传递
- Context 取消机制
- Context 与超时控制
- Context 使用最佳实践

### 3.5 并发模式
- Worker Pool 模式
- Producer-Consumer 模式
- Fan-In/Fan-Out 模式
- Pipeline 模式
- ErrGroup 使用

---

## 第四部分：标准库精讲

### 4.1 常用标准库
- fmt - 格式化 I/O
- strings - 字符串处理
- strconv - 类型转换
- time - 时间处理
- encoding/json - JSON 编解码
- encoding/xml - XML 编解码
- bytes - 字节 slice 操作
- io 与 io/ioutil - I/O 操作
- bufio - 带缓冲的 I/O
- os/os.File - 文件系统操作
- filepath - 路径处理

### 4.2 网络编程
- net 包基础
- TCP/UDP 编程
- http 客户端 (net/http)
- http 服务器
- http 路由与中间件
- http/2 支持

### 4.3 数据库操作
- database/sql 接口
- SQL 驱动使用 (mysql, postgres)
- 连接池配置
- 事务处理
- 预编译语句
- ORM 简介 (GORM, sqlx)

### 4.4 测试与基准测试
- testing 包
- 单元测试编写
- 表格驱动测试
- 测试覆盖率
- 基准测试 (Benchmark)
- 示例测试 (Example)
- Test Main
- .mock 与依赖注入测试

### 4.5 日志与调试
- log 包基础
- 结构化日志
- 日志级别
- 日志输出配置
- pprof 性能分析
- trace 工具

---

## 第五部分：Web 开发

### 5.1 Web 框架
- Gin 框架
  - 路由与分组
  - 中间件
  - 参数绑定与验证
  - 错误处理
  - JSON 响应
- Echo 框架
- Fiber 框架
- 框架对比与选型

### 5.2 中间件开发
- 中间件原理
- 自定义中间件
- CORS 中间件
- 日志中间件
- 认证中间件
- 限流中间件

### 5.3 认证与授权
- Session/Cookie 认证
- JWT Token 认证
- OAuth2.0 集成
- RBAC 权限模型
- Casbin 使用

### 5.4 API 设计
- RESTful API 设计规范
- API 版本管理
- 接口文档 (Swagger/OpenAPI)
- API 限流与熔断
- 统一响应格式

### 5.5 微服务基础
- 微服务架构概念
- gRPC 基础
- Protocol Buffers
- 服务发现
- 负载均衡

---

## 第六部分：工程实践

### 6.1 项目结构
- 标准项目布局 (cmd, pkg, internal)
- 配置管理 (viper)
- 环境变量管理
- 日志配置
- 项目组织最佳实践

### 6.2 代码规范
- Effective Go
- Go 代码风格指南
- 命名规范
- 注释规范
- golint, golangci-lint
- pre-commit hooks

### 6.3 依赖管理
- Go Modules 进阶
- 依赖版本锁定
- 依赖升级策略
- 清理未使用依赖

### 6.4 构建与部署
- 交叉编译
- Docker 容器化
- Dockerfile 编写
- 多阶段构建
- CI/CD 流程 (GitHub Actions, GitLab CI)
- K8s 部署基础

### 6.5 监控与可观测性
- 指标收集 (Prometheus)
- 链路追踪 (Jaeger, OpenTelemetry)
- 健康检查端点
- 结构化日志 (zap, zerolog)
- 告警配置

---

## 第七部分：高级主题

### 7.1 内存管理
- 内存分配机制
- 逃逸分析
- 垃圾回收 (GC) 原理
- 内存优化技巧
- sync.Pool 使用

### 7.2 反射与 Unsafe
- reflect 包基础
- 反射的应用场景
- 反射的性能考虑
- unsafe 包
- 指针转换

### 7.3 汇编基础
- Go 汇编简介
- 查看汇编代码
- 性能优化场景
- 内联函数

### 7.4 设计模式
- 单例模式
- 工厂模式
- 选项模式 (Functional Options)
- 依赖注入
- 观察者模式
- 策略模式

### 7.5 性能优化
- 性能分析方法 (pprof)
- CPU 分析
- 内存分析
- 阻塞分析
- 性能优化案例
- Benchmark 驱动开发

---

## 第八部分：实战项目

### 8.1 入门级项目
- 命令行工具 (CLI)
- Todo List API
- 博客系统后端
- 文件上传服务

### 8.2 进阶级项目
- 分布式任务队列
- API 网关
- 即时通讯服务
- 短链服务

### 8.3 高级项目
- 微服务电商平台
- 实时数据推送系统
- 服务网格 Sidecar
- 自研数据库驱动

### 8.4 开源贡献
- 阅读优秀开源项目源码
- 提交 PR 参与开源项目
- 发布自己的开源库

---

## 学习路线建议

| 阶段 | 内容 | 预计时间 |
|------|------|----------|
| 入门 | 第一部分 + 第二部分 | 2-3 周 |
| 进阶 | 第三部分 + 第四部分 | 3-4 周 |
| 实战 | 第五部分 + 第六部分 | 4-6 周 |
| 提高 | 第七部分 | 4 周 + |
| 精通 | 第八部分实战 | 持续实践 |

---

## 推荐学习资源

### 官方文档
- Go 官方网站：go.dev
- Go 中文文档：golang.org/doc
- Go Blog：go.dev/blog

### 书籍
- 《Go 程序设计语言》(The Go Programming Language)
- 《Go 语言实战》
- 《Go 语言设计与实现》

### 在线资源
- Go by Example
- Uber Go Style Guide
- Golang Design Notes

### 实践平台
- LeetCode (Go 语言刷题)
- Exercism
- Codewars
