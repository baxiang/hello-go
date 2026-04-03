# 02-基础语法

本章节介绍Go语言的基础语法，包括变量、常量、数据类型、运算符和流程控制。

## 学习目标

完成本章节学习后，你将能够：

- 掌握变量和常量的声明与使用
- 理解Go的基本数据类型
- 使用运算符进行计算
- 编写条件判断和循环语句
- 使用defer延迟执行

## 章节内容

### 01-语法基础

**核心内容**：
- 变量与常量声明
- 基本数据类型（int、float、string、bool）
- 运算符与表达式
- 格式化输入输出
- 注释规范

**学习时间**：1-2天

**实践要求**：
- [ ] 编写变量声明代码
- [ ] 使用不同数据类型
- [ ] 练习运算符使用
- [ ] 编写格式化输出程序

### 02-流程控制

**核心内容**：
- if/else条件语句
- switch语句（包括type switch）
- for循环（三种形式）
- break、continue、goto
- defer语句

**学习时间**：1-2天

**实践要求**：
- [ ] 编写条件判断程序
- [ ] 使用switch实现多分支
- [ ] 练习三种for循环
- [ ] 理解defer执行顺序

## 前置知识

- 完成01-环境与工具章节
- 理解基本的编程概念
- 至少一门编程语言基础（有帮助但非必需）

## 学习重点

### 变量声明

```go
// 短变量声明（最常用）
name := "Go"

// 标准声明
var age int = 13

// 类型推断
var version = "1.21"
```

### 常量与iota

```go
const (
    StatusOK = 200
    StatusNotFound = 404
)

const (
    Sunday = iota  // 0
    Monday         // 1
    Tuesday        // 2
)
```

### for循环

```go
// 标准for循环
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// 类似while
i := 0
for i < 10 {
    fmt.Println(i)
    i++
}

// 无限循环
for {
    // do something
    if condition {
        break
    }
}
```

### defer使用

```go
func example() {
    defer fmt.Println("deferred")  // 最后执行
    fmt.Println("normal")
    // 输出: normal, deferred
}
```

## 检查清单

完成本章节后，验证你的掌握程度：

- [ ] 能够正确声明变量和常量
- [ ] 理解不同数据类型的用途
- [ ] 熟练使用运算符
- [ ] 掌握if/switch条件判断
- [ ] 熟练使用三种for循环
- [ ] 理解defer的执行时机

## 常见陷阱

### 1. 变量遮蔽

```go
var x = 10
if true {
    x := 20  // 新变量，遮蔽外部x
    fmt.Println(x)  // 20
}
fmt.Println(x)  // 10
```

### 2. defer在循环中使用

```go
// ❌ 错误：所有defer都在函数结束时执行
for i := 0; i < 3; i++ {
    defer fmt.Println(i)  // 输出: 2, 1, 0
}

// ✅ 正确：立即执行
for i := 0; i < 3; i++ {
    defer func(i int) {
        fmt.Println(i)
    }(i)
}
```

## 练习建议

1. 编写一个计算器程序（加减乘除）
2. 实现一个猜数字游戏
3. 编写九九乘法表
4. 实现简单的登录验证

## 延伸阅读

- [Go语言规范](https://golang.org/ref/spec)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go语言入门教程](https://tour.golang.org/)