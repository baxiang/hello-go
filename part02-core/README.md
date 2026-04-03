# Part 02: 核心特性

本部分深入讲解Go语言的核心特性，包括结构体、接口、错误处理等Go语言最重要的设计理念。

## 学习目标

完成本部分学习后，你将能够：

- 掌握结构体和方法的定义
- 理解接口和鸭子类型
- 实现优雅的错误处理
- 使用Go Modules管理依赖
- 应用泛型解决实际问题

## 章节结构

### 01-结构体与方法

**学习时间**: 2-3天

**核心内容**:
- 结构体定义与实例化
- 匿名字段与结构体嵌入
- 结构体标签（struct tags）
- 方法定义（值接收者 vs 指针接收者）
- Stringer接口实现

**实践要求**:
- [ ] 定义和使用结构体
- [ ] 实现方法
- [ ] 使用结构体标签
- [ ] 理解嵌入和组合

### 02-接口

**学习时间**: 2-3天

**核心内容**:
- 接口定义与实现
- 空接口interface{}
- 类型断言与类型开关
- 接口嵌套
- 接口设计原则

**实践要求**:
- [ ] 定义和实现接口
- [ ] 使用类型断言
- [ ] 理解接口的隐式实现
- [ ] 设计合理的接口

### 03-错误处理

**学习时间**: 2-3天

**核心内容**:
- error接口
- 自定义错误类型
- errors包（Is、As、Join、Unwrap）
- 错误包装（Wrap/Wrapf）
- panic与recover
- 错误处理最佳实践

**实践要求**:
- [ ] 创建自定义错误
- [ ] 使用错误包装
- [ ] 处理panic和recover
- [ ] 实现优雅的错误处理

### 04-包管理

**学习时间**: 1-2天

**核心内容**:
- Go Modules详解
- go.mod与go.sum解析
- 依赖版本管理
- 私有仓库配置
- 发布自己的包

**实践要求**:
- [ ] 创建Go Module
- [ ] 管理依赖版本
- [ ] 发布包到GitHub

### 05-泛型

**学习时间**: 2-3天

**核心内容**:
- 泛型语法基础
- 类型参数与类型约束
- 泛型函数与泛型类型
- 标准库中的泛型
- 泛型使用场景与限制

**实践要求**:
- [ ] 编写泛型函数
- [ ] 定义泛型类型
- [ ] 理解类型约束
- [ ] 应用泛型解决实际问题

## 前置知识

- 完成**Part 01: Go语言基础**
- 理解基础语法和数据结构
- 掌握函数和指针

## 核心概念

### 接口与多态

```go
// 接口定义
type Writer interface {
    Write([]byte) (int, error)
}

// 隐式实现
type FileLogger struct{}

func (f *FileLogger) Write(data []byte) (int, error) {
    // 实现Write方法，自动实现Writer接口
    return len(data), nil
}

// 多态使用
func Log(w Writer, msg string) {
    w.Write([]byte(msg))
}

// 任何实现了Write方法的类型都可以使用
log(&FileLogger{}, "hello")
```

### 错误处理

```go
// 自定义错误
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// 错误包装
func processFile(path string) error {
    if err := validatePath(path); err != nil {
        return fmt.Errorf("process file: %w", err)  // 包装错误
    }
    return nil
}

// 错误检查
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // 处理特定错误类型
}
```

### 泛型示例

```go
// 泛型函数
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// 使用
minInt := Min(1, 2)        // int
minFloat := Min(1.5, 2.5)  // float64

// 泛型类型
type Stack[T any] struct {
    elements []T
}

func (s *Stack[T]) Push(v T) {
    s.elements = append(s.elements, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.elements) == 0 {
        var zero T
        return zero, false
    }
    v := s.elements[len(s.elements)-1]
    s.elements = s.elements[:len(s.elements)-1]
    return v, true
}
```

## 学习重点

### 1. 接口的隐式实现

```go
// Go的接口是隐式实现的
type Reader interface {
    Read(p []byte) (n int, err error)
}

// 只要实现了Read方法，就实现了Reader接口
type MyReader struct{}

func (r *MyReader) Read(p []byte) (n int, err error) {
    return 0, nil
}

// 可以赋值给接口类型
var reader Reader = &MyReader{}
```

### 2. 接收者选择

```go
type Counter struct {
    value int
}

// 值接收者：不修改原对象
func (c Counter) Value() int {
    return c.value
}

// 指针接收者：可以修改原对象
func (c *Counter) Increment() {
    c.value++
}

// 选择原则：
// 1. 需要修改对象 → 指针接收者
// 2. 对象很大 → 指针接收者（避免拷贝）
// 3. 一致性 → 如果有指针接收者，都用指针接收者
```

### 3. 错误处理最佳实践

```go
// ✅ 好的错误处理
func Process(id int) error {
    user, err := getUser(id)
    if err != nil {
        return fmt.Errorf("get user: %w", err)
    }
    
    if err := validate(user); err != nil {
        return fmt.Errorf("validate user: %w", err)
    }
    
    return nil
}

// ❌ 避免
func Process(id int) error {
    user, err := getUser(id)
    if err != nil {
        return err  // 缺少上下文
    }
    
    // ... 没有 defer 清理资源
}
```

## 检查清单

完成本部分后，验证你的掌握程度：

- [ ] 理解结构体和方法的定义
- [ ] 掌握值接收者和指针接收者的区别
- [ ] 理解接口的隐式实现
- [ ] 能够设计合理的接口
- [ ] 掌握错误处理的最佳实践
- [ ] 理解Go Modules的工作原理
- [ ] 能够使用泛型解决实际问题

## 常见陷阱

### 1. 接口nil判断

```go
var r Reader  // r是nil
fmt.Println(r == nil)  // true

var r Reader
r = (*MyReader)(nil)  // r不是nil！
fmt.Println(r == nil)  // false（接口有类型信息）
```

### 2. 值接收者修改无效

```go
type Counter struct {
    value int
}

// ❌ 值接收者无法修改
func (c Counter) Increment() {
    c.value++  // 修改的是副本
}

// ✅ 使用指针接收者
func (c *Counter) Increment() {
    c.value++  // 修改原对象
}
```

### 3. 忽略错误

```go
// ❌ 忽略错误
data, _ := ioutil.ReadFile("file.txt")

// ✅ 处理错误
data, err := ioutil.ReadFile("file.txt")
if err != nil {
    return fmt.Errorf("read file: %w", err)
}
```

## 实践项目

1. **日志系统**
   - 使用接口实现多种日志输出
   - 错误处理和包装

2. **泛型容器**
   - 实现泛型的栈、队列
   - 泛型排序函数

3. **包管理实践**
   - 创建并发布自己的包
   - 管理依赖版本

## 延伸阅读

- [Effective Go - Interfaces](https://golang.org/doc/effective_go#interfaces)
- [Go错误处理](https://blog.golang.org/error-handling-and-go)
- [Go Modules教程](https://blog.golang.org/using-go-modules)
- [Go泛型](https://go.dev/blog/intro-generics)

## 下一步学习

完成Part 02后，建议继续学习：

- **Part 03: 并发编程** - Goroutine、Channel、Context
- **Part 04: 标准库** - 常用标准库详解
- **Part 05: Web开发** - HTTP服务、框架使用

---

## 总结

Part 02涵盖了Go语言最核心的特性：

**重点内容**:
1. ✅ 结构体与方法的定义和使用
2. ✅ 接口的隐式实现和设计原则
3. ✅ 错误处理的最佳实践
4. ✅ Go Modules依赖管理
5. ✅ 泛型的使用场景

**学习时间**: 约2-3周（每天2-3小时）

**学习建议**: 
- 理解Go的设计哲学
- 多写代码实践接口和错误处理
- 掌握泛型的使用场景和限制

这些是Go语言最重要的特性，务必深入理解！