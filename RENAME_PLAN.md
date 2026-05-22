# 文件重命名方案

## 第一阶段：修复文件命名规范

### 需要重命名的文件清单（27个）

#### Part 03 Concurrency（11个文件）- 最混乱

**当前状态**:
```
01-Goroutine 深度解析与 GMP 模型.md  ❌ 中文空格
02-1-Select 语句深度解析.md          ❌ 子编号+中文空格
02-2-Select 底层原理深度解析.md      ❌ 子编号+中文空格
02-3-Select 通俗讲解.md              ❌ 子编号+中文空格
02-4-Select 企业生产级实战.md        ❌ 子编号+中文空格
02-Channel 高级用法与模式.md         ❌ 编号冲突+中文空格
03-同步原语详解.md                   ✅ OK
04-Context 深度解析.md               ❌ 中文空格
05-并发模式与最佳实践.md             ❌ 中文空格
06-内存模型与竞态检测.md             ❌ 中文空格
README.md                            ✅ OK
```

**重命名方案**:
```
01-Goroutine 深度解析与 GMP 模型.md  -> 01-Goroutine与GMP模型.md
02-Channel 高级用法与模式.md         -> 02-Channel详解.md
02-1-Select 语句深度解析.md          -> 03-Select基础用法.md
02-2-Select 底层原理深度解析.md      -> 04-Select底层原理.md
02-3-Select 通俗讲解.md              -> (内容合并到03-Select基础用法.md)
02-4-Select 企业生产级实战.md        -> 05-Select生产实战.md
03-同步原语详解.md                   -> 06-同步原语详解.md
04-Context 深度解析.md               -> 07-Context详解.md
05-并发模式与最佳实践.md             -> 08-并发模式实战.md
06-内存模型与竞态检测.md             -> 09-内存模型与竞态.md
```

#### Part 01 Basics（1个）

```
09-Cobra 命令行工具.md -> 09-Cobra命令行工具.md
```

#### Part 04 Stdlib（3个）

```
06-文件与 IO 操作.md      -> 06-文件与IO操作.md
09-Viper 配置管理.md      -> 09-Viper配置管理.md
10-GORM 数据库操作.md     -> 10-GORM数据库操作.md
```

#### Part 05 Web（4个）

```
01-Web 框架.md            -> 01-Web框架详解.md
04-API 设计.md            -> 04-API设计规范.md
06-WebSocket 实时通信.md  -> 06-WebSocket实时通信.md
07-GraphQL 基础.md        -> 07-GraphQL基础.md
```

#### Part 06 Engineering（1个）

```
08-Wire 依赖注入.md       -> 08-Wire依赖注入.md
```

#### Part 07 Advanced（1个）

```
02-反射与 Unsafe.md       -> 02-反射与Unsafe详解.md
```

#### Part 08 Projects（1个）

```
05-Cloud Native 项目.md   -> 05-CloudNative项目.md
```

#### Part 08 Projects 子目录（7个）

**go-zero/**:
```
01-go-zero 核心概念与基础.md  -> 01-go-zero核心概念.md
02-go-zero 进阶与项目.md      -> 02-go-zero进阶实战.md
```

**nats/**:
```
01-NATS 核心概念与基础.md                      -> 01-NATS核心概念.md
02-NATS JetStream 持久化.md                   -> 02-NATS-JetStream持久化.md
03-入门项目-NATS 消息订阅发布系统.md           -> 03-NATS消息订阅发布实战.md
04-进阶项目-JetStream 订单处理系统.md         -> 04-JetStream订单处理实战.md
```

**temporal/**:
```
01-Temporal 核心概念与基础.md  -> 01-Temporal核心概念.md
02-Temporal Go 客户端与使用.md -> 02-Temporal-Go客户端实战.md
03-Temporal 工作流模式.md     -> 03-Temporal工作流模式.md
```

## 执行步骤

### 步骤1: Part 03 Concurrency（最复杂）

```bash
# 创建备份
cd /Users/baxiang/Documents/hello-go/02-进阶编程
tar -czf backup.tar.gz *.md

# 执行重命名
mv "01-Goroutine 深度解析与 GMP 模型.md" "01-Goroutine与GMP模型.md"
mv "02-Channel 高级用法与模式.md" "02-Channel详解.md"
mv "02-1-Select 语句深度解析.md" "03-Select基础用法.md"
mv "02-2-Select 底层原理深度解析.md" "04-Select底层原理.md"
# mv "02-3-Select 通俗讲解.md" 内容合并到03
mv "02-4-Select 企业生产级实战.md" "05-Select生产实战.md"
mv "03-同步原语详解.md" "06-同步原语详解.md"
mv "04-Context 深度解析.md" "07-Context详解.md"
mv "05-并发模式与最佳实践.md" "08-并发模式实战.md"
mv "06-内存模型与竞态检测.md" "09-内存模型与竞态.md"

# 合并通俗讲解内容到基础用法
cat "02-3-Select 通俗讲解.md" >> "03-Select基础用法.md"
rm "02-3-Select 通俗讲解.md"
```

### 步骤2-8: 其他部分（批量处理）

将使用批量脚本处理其他部分的文件重命名。

## 注意事项

1. **Git影响**: 重命名后需要使用 `git mv` 保持Git历史
2. **引用更新**: 需要更新所有文件中的内部引用
3. **Roadmap更新**: 需要更新 go-learning-roadmap.md 中的文件名引用
4. **验证**: 重命名后需要验证所有文件可访问

## 后续任务

完成文件重命名后，将继续：
- 创建 Part 09 云原生章节
- 创建 Part 10 性能调优章节
- 补充 AI 辅助开发内容
- 增强工业实践案例