# 1. Goroutine 深度解析与 GMP 模型

> 理解 Goroutine 的本质和 Go 调度器的工作原理，是掌握 Go 并发的基石。

## 1.1 什么是 Goroutine？

### 1.1.1 概念定义

Goroutine 是 Go 语言特有的并发执行单元，是由 Go 运行时（runtime）管理的轻量级线程。

**Goroutine vs 操作系统线程对比：**

| 特性 | 操作系统线程 | Goroutine |
|------|-------------|-----------|
| 栈大小 | 固定 1-8MB | 初始 2KB，可动态增长 |
| 创建成本 | 高 (~1ms) | 极低 (~2KB 内存) |
| 切换成本 | 高 (~1-2μs) | 极低 (~0.2μs) |
| 调度方式 | 操作系统调度 | Go 运行时调度 |
| 数量限制 | 数千 | 数十万 |

### 1.1.2 为什么 Goroutine 如此高效？

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    // 启动 100 万个 goroutine
    const count = 1000000
    start := time.Now()
    
    for i := 0; i < count; i++ {
        go func() {
            // 什么都不做
        }()
    }
    
    // 等待所有 goroutine 启动
    runtime.Gosched()
    
    fmt.Printf("创建 %d 个 goroutine 耗时：%v\n", count, time.Since(start))
    fmt.Printf("当前 goroutine 数量：%d\n", runtime.NumGoroutine())
}
```

**输出示例：**
```
创建 1000000 个 goroutine 耗时：98ms
当前 goroutine 数量：1000003
```

### 1.1.3 Goroutine 的生命周期

```
创建 (GNEW) → 就绪 (_Grunnable) → 运行 (GRUNNING) → 等待 (Gwaiting) → 再次就绪 → 退出 (Gdead)
```

## 1.2 GMP 模型详解

### 1.2.1 GMP 是什么？

**GMP 是 Go 运行时调度器的核心模型：**

```
G (Goroutine)     M (Machine)      P (Processor)
┌───────────┐    ┌──────────┐    ┌──────────┐
│ 执行的    │    │  系统    │    │  调度    │
│ 代码      │    │  线程    │    │  上下文  │
│           │    │          │    │          │
│ 栈：2KB+  │    │  栈：2MB+│    │ 本地队列  │
│           │    │          │    │          │
│ 状态：    │    │  状态：  │    │  数量：  │
│ runnable  │    │ running  │    │ = CPU 核  │
│ waiting   │    │ sleeping │    │  心数    │
└───────────┘    └──────────┘    └──────────┘

关系:
- 1 个 P 最多运行 1 个 M
- 1 个 M 需要绑定 1 个 P 才能运行
- P 的本地队列可存放最多 256 个 G
- 全局队列存放所有 P 共享的 G
```

### 1.2.2 调度流程

```
┌─────────┐    ┌─────────┐    ┌─────────┐
│ 全局队列 │◄──►│ P 本地队列│◄──►│   M    │
│(G 队列)  │    │(256 个 G) │    │(执行)  │
└────┬────┘    └────┬────┘    └────┬────┘
     │             │             │
     ▼             ▼             ▼
┌─────────────────────────────────────────┐
│         Work-Stealing 策略              │
│  1. 从本地队列取 G                      │
│  2. 从全局队列取 G                      │
│  3. 从网络轮询器取 G                    │
│  4. 从其他 P 偷取 G                     │
└─────────────────────────────────────────┘
```

### 1.2.3 调度时机

Go 调度器在以下情况会进行调度：

1. **主动让出** (`runtime.Gosched()`)
2. **阻塞等待** (channel、system call)
3. **定时器触发**
4. **GC 完成后**
5. **内存分配过多**

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    // 1. 主动让出
    go func() {
        for i := 0; i < 100; i++ {
            if i % 10 == 0 {
                runtime.Gosched()  // 主动让出 CPU
                fmt.Println("goroutine A 主动让出")
            }
        }
    }()
    
    // 2. 阻塞等待
    ch := make(chan int)
    go func() { ch <- 1 }()
    <-ch
    
    // 3. 定时器触发
    time.AfterFunc(time.Second, func() {
        fmt.Println("定时器触发")
    })
}
```

## 1.3 实际案例：Kubernetes 中的 Goroutine 使用

### 1.3.1 Kubernetes Informer 模式

Kubernetes 使用大量 goroutine 处理资源监听：

```go
// k8s.io/client-go/tools/cache/controller.go (简化)

// New 创建 Controller
func New(config *Config) Controller {
    c := &controller{
        queue:    NewFIFO(),
        handlers: config.Handlers,
    }
    
    // 启动多个 worker goroutine
    for i := 0; i < config.Workers; i++ {
        go c.worker(i)
    }
    
    return c
}

// worker 处理队列中的任务
func (c *controller) worker(workerID int) {
    for {
        // 从队列获取任务
        item, quit := c.queue.Get()
        if quit {
            return
        }
        
        // 处理任务
        if err := c.processFunc(item); err != nil {
            // 错误处理
        }
        
        // 标记完成
        c.queue.Done(item)
    }
}
```

**Kubernetes 并发规模：**

| 组件 | Goroutine 用途 | 规模 |
|------|---------------|------|
| Kubelet | 状态同步、Pod 管理、健康检查 | 数千 |
| Controller Manager | Deployment、ReplicaSet、Node 等 Controller | 数百 |
| API Server | 请求处理、Watch、Etcd 访问 | 数万 |

### 1.3.2 etcd 中的 Goroutine 使用

```go
// etcdserver/server.go (简化)

// StartEtcd 启动 etcd 服务器
func (s *EtcdServer) StartEtcd() error {
    // 启动 raft 消息处理 goroutine
    for i := 0; i < s.cfg.NumOfMessagesToApply; i++ {
        go func() {
            for {
                select {
                case <-s.msgAppC:
                    // 处理 raft 消息
                case <-s.stopC:
                    return
                }
            }
        }()
    }
    
    // 启动 snapshot goroutine
    go func() {
        for {
            select {
            case <-s.snapC:
                // 保存快照
            case <-s.stopC:
                return
            }
        }
    }()
    
    // 启动租约检查 goroutine
    go func() {
        for {
            select {
            case <-time.After(time.Second):
                // 检查过期租约
            case <-s.stopC:
                return
            }
        }
    }()
    
    return nil
}
```

## 1.4 Goroutine 泄漏与排查

### 1.4.1 常见的 Goroutine 泄漏

```go
// 泄漏示例 1: channel 发送后无人接收
func leak1() {
    ch := make(chan int)
    go func() {
        ch <- 1  // 阻塞，永远不会有人接收
    }()
    // 泄漏！
}

// 泄漏示例 2: channel 接收后无人发送
func leak2() {
    ch := make(chan int)
    go func() {
        <-ch  // 阻塞，永远不会有人发送
    }()
    // 泄漏！
}

// 正确示例：使用 context 取消
func correct() {
    ctx, cancel := context.WithCancel(context.Background())
    
    go func() {
        select {
        case <-ctx.Done():
            return  // 正确取消
        }
    }()
    
    cancel()  // 取消 goroutine
}
```

### 1.4.2 使用 pprof 排查

```go
package main

import (
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)

func main() {
    // 启动 pprof
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    
    // 模拟泄漏
    go func() {
        ch := make(chan int)
        for {
            go func() { ch <- 1 }()
            time.Sleep(time.Millisecond)
        }
    }()
    
    // 打印 goroutine 数量
    for {
        fmt.Printf("Goroutine 数量：%d\n", runtime.NumGoroutine())
        time.Sleep(time.Second)
    }
}
```

```bash
# 查看 goroutine 堆栈
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 查看 top goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## 1.5 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────┤
│ ✓ Goroutine 是 Go 并发的核心                                │
│   └── 轻量级、栈可增长、由运行时管理                        │
│                                                             │
│ ✓ GMP 模型是调度器的核心                                   │
│   └── G(goroutine) / M(线程) / P(处理器)                  │
│                                                             │
│ ✓ Work-Stealing 策略实现负载均衡                          │
│   └── 本地队列 → 全局队列 → 偷取                          │
│                                                             │
│ ✓ 网络轮询器实现非阻塞 I/O                                │
│   └── epoll/kqueue 集成到调度器                           │
│                                                             │
│ ✓ 著名项目大量使用 goroutine                              │
│   └── Kubernetes、etcd、Docker 等                         │
│                                                             │
│ ✓ 注意 goroutine 泄漏问题                                │
│   └── 使用 pprof 排查                                     │
└─────────────────────────────────────────────────────────────┘
```

---

[下一章：Channel 高级用法与模式 →](./02-Channel 高级用法与模式.md)