# 06-标准库基础

本章节介绍Go语言标准库的基础使用，包括时间处理和字符串操作。

## 学习目标

完成本章节学习后，你将能够：

- 熟练使用time包处理时间
- 掌握strings和bytes包的常用操作
- 理解string的不可变性和底层原理
- 进行高效的字符串处理

## 章节内容

### 01-时间处理

**核心内容**：
- time.Time结构体
- 时间格式化与解析
- 时间运算（Add、Sub、After、Before）
- 定时器（Timer、Ticker）
- 时区处理
- 时间戳转换
- 耗时计算（Since、Until）

**学习时间**：2-3天

**实践要求**：
- [ ] 创建和格式化时间
- [ ] 计算时间差
- [ ] 实现定时任务
- [ ] 处理不同时区

### 02-字符串与字节

**核心内容**：
- string不可变性
- strings包常用函数
- bytes包与字节操作
- strconv类型转换
- 字符串性能优化
- 编码转换（base64、hex）

**学习时间**：2-3天

**实践要求**：
- [ ] 使用strings包处理字符串
- [ ] 操作字节切片
- [ ] 类型转换
- [ ] 字符串拼接优化

## 前置知识

- 完成02-基础语法章节
- 完成03-数据结构章节
- 理解切片和数组

## 核心概念

### 时间处理基础

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // 获取当前时间
    now := time.Now()
    fmt.Println("当前时间:", now)
    
    // 时间格式化（Go的特殊格式）
    fmt.Println(now.Format("2006-01-02 15:04:05"))
    
    // 时间运算
    future := now.Add(24 * time.Hour)
    fmt.Println("明天此时:", future)
    
    // 计算时间差
    duration := future.Sub(now)
    fmt.Println("相差:", duration)
    
    // 定时器
    timer := time.NewTimer(1 * time.Second)
    <-timer.C
    fmt.Println("1秒后")
}
```

### 字符串基础

```go
package main

import (
    "fmt"
    "strings"
    "bytes"
)

func main() {
    // 字符串不可变
    s := "hello"
    // s[0] = 'H'  // 编译错误！
    
    // strings包常用函数
    fmt.Println(strings.Contains(s, "ell"))  // true
    fmt.Println(strings.ToUpper(s))           // HELLO
    fmt.Println(strings.Split("a,b,c", ",")) // [a b c]
    
    // 字符串拼接优化
    var builder strings.Builder
    for i := 0; i < 100; i++ {
        builder.WriteString("x")
    }
    result := builder.String()
    
    // bytes包
    var buf bytes.Buffer
    buf.WriteString("hello")
    buf.WriteString(" ")
    buf.WriteString("world")
    fmt.Println(buf.String())
}
```

## 学习重点

### 1. 时间格式化

Go使用特殊的参考时间进行格式化：

```go
// 参考时间：Mon Jan 2 15:04:05 MST 2006
// 格式化字符串使用这个时间的各个部分

now := time.Now()

// 常用格式
now.Format("2006-01-02")              // 2024-01-15
now.Format("2006-01-02 15:04:05")      // 2024-01-15 10:30:00
now.Format("2006-01-02 15:04:05.000")  // 带毫秒
now.Format("2006-01-02 15:04:05 -0700") // 带时区

// 解析时间
t, _ := time.Parse("2006-01-02", "2024-01-15")
```

### 2. 字符串性能优化

```go
// ❌ 频繁分配（性能差）
s := ""
for i := 0; i < 1000; i++ {
    s += "x"  // 每次都创建新字符串
}

// ✅ 使用strings.Builder（推荐）
var builder strings.Builder
builder.Grow(1000)  // 预分配
for i := 0; i < 1000; i++ {
    builder.WriteString("x")
}
s := builder.String()

// ✅ 使用bytes.Buffer
var buf bytes.Buffer
buf.Grow(1000)
for i := 0; i < 1000; i++ {
    buf.WriteString("x")
}
s := buf.String()
```

### 3. 类型转换

```go
import "strconv"

// 字符串转数字
n, _ := strconv.Atoi("123")
f, _ := strconv.ParseFloat("3.14", 64)

// 数字转字符串
s := strconv.Itoa(123)
s := strconv.FormatFloat(3.14, 'f', 2, 64)

// 其他类型
b, _ := strconv.ParseBool("true")
s := strconv.FormatBool(true)
```

## 检查清单

完成本章节后，验证你的掌握程度：

- [ ] 理解时间格式化的特殊规则
- [ ] 能够进行时间运算
- [ ] 实现过定时器和计时器
- [ ] 掌握strings包的常用函数
- [ ] 理解string的不可变性
- [ ] 掌握字符串拼接的性能优化
- [ ] 熟练使用strconv进行类型转换

## 常见陷阱

### 1. 时间格式化错误

```go
// ❌ 错误：使用常见格式
now.Format("YYYY-MM-DD")  // 错误！

// ✅ 正确：使用Go的参考时间
now.Format("2006-01-02")  // 正确
```

### 2. 字符串拼接性能问题

```go
// ❌ 在循环中拼接
result := ""
for _, s := range items {
    result += s  // O(n²) 复杂度
}

// ✅ 使用Builder
var builder strings.Builder
for _, s := range items {
    builder.WriteString(s)  // O(n) 复杂度
}
result := builder.String()
```

### 3. 时区问题

```go
// 注意时区
now := time.Now()           // 本地时间
utc := time.Now().UTC()     // UTC时间

// 加载时区
loc, _ := time.LoadLocation("America/New_York")
ny := now.In(loc)
```

## 实践项目

1. **时间工具库**
   - 实现时间格式化工具
   - 计算工作日
   - 时间范围判断

2. **字符串处理工具**
   - 字符串统计工具
   - 文本格式化工具
   - 日志解析器

3. **性能对比测试**
   - 对比不同字符串拼接方法
   - 编写benchmark测试

## 性能优化技巧

### 1. 字符串拼接

```go
// 性能从低到高：
// 1. + 拼接（少量时可用）
s := s1 + s2

// 2. fmt.Sprintf（灵活但较慢）
s := fmt.Sprintf("%s%s", s1, s2)

// 3. strings.Builder（推荐）
var builder strings.Builder
builder.WriteString(s1)
builder.WriteString(s2)
s := builder.String()

// 4. 预分配容量（最优）
builder.Grow(len(s1) + len(s2))
```

### 2. 避免不必要的转换

```go
// ❌ 频繁转换
for _, b := range []byte(s) {
    // 处理字节
}

// ✅ 直接处理
for i := 0; i < len(s); i++ {
    b := s[i]  // 不需要转换
}
```

## 延伸阅读

- [Go时间处理文档](https://golang.org/pkg/time/)
- [strings包文档](https://golang.org/pkg/strings/)
- [bytes包文档](https://golang.org/pkg/bytes/)
- [strconv包文档](https://golang.org/pkg/strconv/)
- [Go字符串优化](https://blog.golang.org/strings)