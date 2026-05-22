# 04-函数与指针

本章节介绍Go语言的函数定义、指针使用以及相关高级特性。

## 学习目标

完成本章节学习后，你将能够：

- 定义和调用函数
- 使用多返回值和可变参数
- 理解匿名函数和闭包
- 掌握指针的概念和使用
- 理解值传递和引用传递

## 章节内容

### 01-函数

**核心内容**：
- 函数定义与调用
- 多返回值
- 可变参数
- 匿名函数
- 闭包与作用域
- init函数详解
- defer执行顺序陷阱

**学习时间**：2-3天

**实践要求**：
- [ ] 编写多种形式的函数
- [ ] 使用多返回值处理错误
- [ ] 实现可变参数函数
- [ ] 理解闭包的捕获机制

### 02-指针

**核心内容**：
- 指针基础概念
- new和make的区别
- 指针与函数参数传递
- 指针的安全性
- 指针与数组/切片

**学习时间**：1-2天

**实践要求**：
- [ ] 理解指针的基本操作
- [ ] 使用指针修改函数外部变量
- [ ] 理解new和make的区别
- [ ] 掌握指针的最佳实践

## 前置知识

- 完成02-基础语法章节
- 完成03-数据结构章节
- 理解变量和数据类型

## 核心概念

### 函数定义

```go
// 基本函数
func add(a, b int) int {
    return a + b
}

// 多返回值
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// 命名返回值
func calc(a, b int) (sum, product int) {
    sum = a + b
    product = a * b
    return  // 自动返回命名变量
}

// 可变参数
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// 使用
sum(1, 2, 3, 4, 5)  // 15
```

### 匿名函数与闭包

```go
// 匿名函数
add := func(a, b int) int {
    return a + b
}
result := add(1, 2)

// 闭包：捕获外部变量
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

c := counter()
fmt.Println(c())  // 1
fmt.Println(c())  // 2
fmt.Println(c())  // 3
```

### 指针基础

```go
// 指针声明和使用
var x int = 10
var p *int = &x

fmt.Println(*p)  // 10 (解引用)
*p = 20
fmt.Println(x)   // 20

// 指针作为函数参数
func increment(p *int) {
    *p++  // 修改指针指向的值
}

x := 10
increment(&x)
fmt.Println(x)  // 11
```

### new vs make

```go
// new: 分配内存，返回指针
p := new(int)  // *int, 值为0

// make: 只用于slice、map、channel
slice := make([]int, 0, 10)
m := make(map[string]int)
ch := make(chan int)

// 区别
// new: 返回指针
// make: 返回初始化后的值（不是指针）
```

## 学习重点

### 1. defer执行顺序

```go
func example() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    fmt.Println("function body")
}
// 输出顺序:
// function body
// 3
// 2
// 1
// LIFO（后进先出）
```

### 2. 闭包捕获变量

```go
// ❌ 常见错误
funcs := []func(){}
for i := 0; i < 3; i++ {
    funcs = append(funcs, func() {
        fmt.Println(i)  // 捕获的是变量i本身
    })
}
for _, f := range funcs {
    f()  // 输出: 3, 3, 3
}

// ✅ 正确做法
for i := 0; i < 3; i++ {
    i := i  // 创建新变量
    funcs = append(funcs, func() {
        fmt.Println(i)
    })
}
// 输出: 0, 1, 2
```

### 3. 值传递 vs 指针传递

```go
// 值传递：拷贝副本
func modifyValue(x int) {
    x = 20  // 不影响原变量
}

// 指针传递：传递地址
func modifyPointer(x *int) {
    *x = 20  // 修改原变量
}

num := 10
modifyValue(num)
fmt.Println(num)  // 10

modifyPointer(&num)
fmt.Println(num)  // 20
```

## 检查清单

完成本章节后，验证你的掌握程度：

- [ ] 能够定义各种形式的函数
- [ ] 熟练使用多返回值
- [ ] 理解可变参数的使用
- [ ] 掌握匿名函数和闭包
- [ ] 理解defer的执行顺序
- [ ] 理解指针的概念
- [ ] 掌握指针作为函数参数
- [ ] 理解new和make的区别

## 最佳实践

### 1. 何时使用指针

**使用指针的场景**：
- 需要修改函数外部变量
- 大结构体，避免拷贝开销
- 需要表示nil（可选参数）
- 实现修改接收者的方法

**不使用指针的场景**：
- 小数据类型（int、float等）
- 只读操作
- 不需要nil表示

### 2. 函数设计原则

```go
// ✅ 好的函数设计
func ReadFile(path string) ([]byte, error) {
    // 单一职责
    // 清晰的错误处理
    // 有意义的名称
}

// ❌ 避免
func doStuff(x interface{}) interface{} {
    // 职责不清
    // 类型不安全
    // 名称不明确
}
```

### 3. defer使用建议

```go
// ✅ 推荐用法：资源清理
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()  // 确保文件关闭

// ✅ 推荐用法：解锁
mu.Lock()
defer mu.Unlock()

// ⚠️ 在循环中谨慎使用
for i := 0; i < 1000; i++ {
    defer cleanup()  // 延迟到函数结束，可能内存泄露
}
```

## 常见陷阱

### 1. defer在循环中

```go
// ❌ 错误：所有defer在函数结束时执行
func processFiles(files []string) error {
    for _, file := range files {
        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close()  // 文件不会立即关闭！
    }
}

// ✅ 正确：使用匿名函数
func processFiles(files []string) error {
    for _, file := range files {
        err := func() error {
            f, err := os.Open(file)
            if err != nil {
                return err
            }
            defer f.Close()  // 在匿名函数结束时关闭
            // 处理文件...
            return nil
        }()
        if err != nil {
            return err
        }
    }
}
```

### 2. nil指针检查

```go
var p *int
fmt.Println(*p)  // panic: nil pointer dereference

// ✅ 检查nil
if p != nil {
    fmt.Println(*p)
}
```

## 练习项目

1. 实现一个简单的计算器（多种运算）
2. 编写一个计数器（使用闭包）
3. 实现链表结构（使用指针）
4. 编写一个简单的栈（使用切片和指针）

## 延伸阅读

- [Go函数](https://golang.org/doc/codewalk/functions/)
- [Go指针](https://golang.org/doc/faq#pass_by_value)
- [defer实现原理](https://blog.golang.org/defer-panic-and-recover)