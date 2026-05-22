# 03-数据结构

本章节介绍Go语言的内置数据结构：数组、切片和Map。

## 学习目标

完成本章节学习后，你将能够：

- 理解数组和切片的区别
- 熟练使用切片的各种操作
- 掌握Map的使用方法
- 理解数据结构的底层原理

## 章节内容

### 01-数组与切片

**核心内容**：
- 数组定义与使用
- 切片的创建与操作
- 切片的扩容机制
- copy和append函数
- 多维数组与切片

**学习时间**：2-3天

**实践要求**：
- [ ] 创建和操作数组
- [ ] 使用切片的各种方法
- [ ] 理解切片扩容原理
- [ ] 实现切片的拷贝和追加

### 02-Map

**核心内容**：
- Map的创建与初始化
- 增删改查操作
- Map的遍历
- Map与切片的组合使用
- sync.Map简介

**学习时间**：1-2天

**实践要求**：
- [ ] 创建和操作Map
- [ ] 实现键值对的增删改查
- [ ] 遍历Map数据
- [ ] 使用Map实现简单缓存

## 前置知识

- 完成02-基础语法章节
- 理解变量和数据类型

## 核心概念

### 数组 vs 切片

```go
// 数组：固定长度
var arr [5]int

// 切片：动态长度
var slice []int
slice = make([]int, 0, 5)  // 长度0，容量5

// 切片更常用
nums := []int{1, 2, 3, 4, 5}
```

### 切片操作

```go
nums := []int{1, 2, 3, 4, 5}

// 切片操作
fmt.Println(nums[1:3])  // [2 3]
fmt.Println(nums[:3])   // [1 2 3]
fmt.Println(nums[2:])   // [3 4 5]

// 追加元素
nums = append(nums, 6, 7)

// 拷贝
dst := make([]int, len(nums))
copy(dst, nums)
```

### Map操作

```go
// 创建Map
m := make(map[string]int)

// 增
m["one"] = 1
m["two"] = 2

// 查
value, exists := m["one"]
if exists {
    fmt.Println(value)  // 1
}

// 删
delete(m, "one")

// 遍历
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}
```

## 学习重点

### 1. 切片扩容机制

```go
// 切片扩容策略
// 小于1024: 2倍扩容
// 大于1024: 1.25倍扩容

slice := make([]int, 0)
for i := 0; i < 10000; i++ {
    slice = append(slice, i)
    // 观察容量变化
    fmt.Printf("len=%d cap=%d\n", len(slice), cap(slice))
}
```

### 2. 切片共享底层数组

```go
// ⚠️ 切片共享底层数组
original := []int{1, 2, 3, 4, 5}
part := original[1:3]  // [2 3]

part[0] = 20
fmt.Println(original)  // [1 20 3 4 5]  原数组被修改！

// ✅ 使用copy避免共享
copied := make([]int, 2)
copy(copied, original[1:3])
```

### 3. Map的并发安全

```go
// ❌ Map不是并发安全的
// 并发读写会panic

// ✅ 使用sync.Map
var m sync.Map
m.Store("key", "value")
value, ok := m.Load("key")
```

## 检查清单

完成本章节后，验证你的掌握程度：

- [ ] 理解数组和切片的区别
- [ ] 熟练使用切片的各种操作
- [ ] 理解切片的扩容机制
- [ ] 掌握Map的增删改查
- [ ] 理解切片共享底层数组的原理
- [ ] 了解sync.Map的使用场景

## 常见陷阱

### 1. 切片引用问题

```go
// ❌ 错误：循环变量地址相同
var slices [][]int
for i := 0; i < 3; i++ {
    slices = append(slices, []int{i})
}
// 可能输出: [[2] [2] [2]]

// ✅ 正确：每次创建新切片
for i := 0; i < 3; i++ {
    s := []int{i}
    slices = append(slices, s)
}
```

### 2. Map的nil检查

```go
var m map[string]int

// ❌ panic: nil map
m["key"] = 1

// ✅ 先初始化
m = make(map[string]int)
m["key"] = 1
```

## 实践项目

1. 实现一个动态数组（类似Java ArrayList）
2. 使用Map实现单词计数器
3. 实现一个简单的缓存系统
4. 编写切片去重函数

## 性能优化技巧

### 1. 预分配容量

```go
// ❌ 频繁扩容
var slice []int
for i := 0; i < 10000; i++ {
    slice = append(slice, i)
}

// ✅ 预分配容量
slice := make([]int, 0, 10000)
for i := 0; i < 10000; i++ {
    slice = append(slice, i)
}
```

### 2. 切片复用

```go
// 复用切片减少内存分配
var buf []byte
for i := 0; i < 100; i++ {
    buf = buf[:0]  // 清空但保留容量
    // 使用buf...
}
```

## 延伸阅读

- [Go切片原理](https://blog.golang.org/go-slices-usage-and-internals)
- [Go Map实现](https://blog.golang.org/maps)
- [Go数据结构](https://research.swtch.com/godata)