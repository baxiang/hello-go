# NATS 百万设备架构分析

## 问题：百万设备场景下的 Subject 设计

### 场景描述

```
设备数量：1,000,000（百万级）
每个设备 ID 不同：device-001, device-002, ..., device-1000000
每个设备需要接收命令：device.{device-id}.command
```

### 核心问题

**Subject 数量本身不是问题，订阅方式才是关键！**

---

## 一、Subject 数量的影响

### 1.1 Subject 本身没有"创建"开销

```
NATS 的 Subject 只是一个字符串，不需要预先创建：

发布时：
  nc.Publish("device.device-001.command", data)  // Subject 即时使用

服务器内部：
  - Subject 不占用内存（除非有订阅者）
  - 没有Subject 注册表
  - 发布到无订阅者的 Subject = 消息丢弃（Core NATS）
```

### 1.2 Subject 数量 ≠ 性能问题

```
发布 100 万个不同 Subject 的消息：

for i := 0; i < 1000000; i++ {
    subject := fmt.Sprintf("device.device-%d.command", i)
    nc.Publish(subject, data)
}

性能影响：
  - 每次发布都是 O(1) 操作
  - Subject 字符串匹配是 O(L)，L = Subject 长度
  - 与 Subject 总数无关
```

---

## 二、订阅方式才是关键

### 2.1 错误方式：为每个设备创建独立订阅

```go
// ❌ 错误：百万订阅 = 内存爆炸
for _, deviceID := range deviceIDs {  // 1,000,000 个设备
    subject := fmt.Sprintf("device.%s.command", deviceID)
    nc.Subscribe(subject, handler)  // 每个设备一个订阅
}

问题：
  1. 每个订阅占用服务器内存（约 1-2KB）
  2. 百万订阅 ≈ 1-2GB 内存
  3. 路由表膨胀，匹配效率下降
  4. 客户端也需要维护百万订阅状态
```

### 2.2 正确方式：使用通配符订阅

```go
// ✅ 正确：一个订阅覆盖所有设备
nc.Subscribe("device.*.command", func(msg *nats.Msg) {
    // 从 Subject 解析设备 ID
    // device.device-001.command → device-001
    parts := strings.Split(msg.Subject, ".")
    deviceID := parts[1]
    
    // 根据设备 ID 分发处理
    dispatchToDevice(deviceID, msg.Data)
})

优势：
  1. 只需 1 个订阅
  2. 内存占用极小
  3. 路由表只有 1 条记录
  4. 匹配效率 O(L)，与设备数无关
```

---

## 三、架构设计对比

### 3.1 场景一：设备端订阅（当前设计）

```
设备端（每个设备一个进程）：
  device-001 → 连接 NATS → 订阅 "device.device-001.command"
  device-002 → 连接 NATS → 订阅 "device.device-002.command"
  ...

分析：
  ✅ 这是正确的！
  - 每个设备只创建 1 个订阅
  - 百万设备 = 百万连接 + 百万订阅
  - 但分布在百万个进程中，每个进程只有 1 个订阅
  - NATS 服务器支持百万连接（单节点 100 万+ 连接）
```

### 3.2 场景二：Hub/服务端订阅

```
Hub 服务（需要向设备发送命令）：

❌ 错误方式：
  Hub 订阅所有设备的响应：
  for _, deviceID := range deviceIDs {
      nc.Subscribe(fmt.Sprintf("device.%s.response", deviceID), handler)
  }
  → 百万订阅在一个进程中 = 内存爆炸

✅ 正确方式：
  方案 A：通配符订阅
  nc.Subscribe("device.*.response", handler)
  
  方案 B：使用 Reply-To（Request/Reply 模式）
  // 发送命令时使用 Reply-To
  msg := &nats.Msg{
      Subject: fmt.Sprintf("device.%s.command", deviceID),
      Reply:   nats.NewInbox(),  // 临时回复地址
      Data:    commandData,
  }
  resp, _ := nc.RequestMsg(msg, timeout)
  // 不需要预先订阅，Request 自动处理
```

---

## 四、性能基准

### 4.1 NATS 服务器性能

```
单节点 NATS Server 性能（官方数据）：

连接数：
  - 支持 1,000,000+ 并发连接
  - 每连接内存开销：约 1-2KB（不含消息缓冲）

订阅数：
  - 理论上限：取决于内存
  - 实际建议：单进程订阅数 < 10,000
  - 使用通配符可将百万订阅压缩为 1 个

吞吐量：
  - Core NATS：18,000,000 msg/s
  - JetStream：3,000,000 msg/s（持久化）
```

### 4.2 订阅方式对比

| 方式 | 订阅数 | 内存占用 | 匹配效率 | 适用场景 |
|------|--------|----------|----------|----------|
| 独立订阅 | 1,000,000 | 1-2GB | O(N) | ❌ 不推荐 |
| 通配符订阅 | 1 | 几 KB | O(L) | ✅ 推荐 |
| Queue Group | N（分散到多实例） | 分散 | O(L) | ✅ 推荐 |

---

## 五、最佳实践

### 5.1 Subject 设计

```
推荐设计：

device.{device-id}.command      # 设备接收命令
device.{device-id}.response     # 设备响应
device.{device-id}.telemetry    # 设备遥测数据
device.{device-id}.status       # 设备状态

通配符订阅：
  device.*.command     → 接收所有设备的命令
  device.*.telemetry   → 接收所有设备的遥测
  device.>             → 接收设备相关的所有消息
```

### 5.2 设备端代码

```go
// 设备端：只订阅自己的 Subject（正确）
deviceID := "device-001"
subject := fmt.Sprintf("device.%s.command", deviceID)

nc.Subscribe(subject, func(msg *nats.Msg) {
    // 处理命令
    handleCommand(msg.Data)
    
    // 响应（使用 msg.Reply）
    if msg.Reply != "" {
        nc.Publish(msg.Reply, responseData)
    }
})

// 每个设备进程只有 1 个订阅，完全没问题
```

### 5.3 Hub/服务端代码

```go
// Hub 服务端：向设备发送命令

// 方式 1：Request/Reply（推荐）
func sendCommand(deviceID string, command []byte) ([]byte, error) {
    subject := fmt.Sprintf("device.%s.command", deviceID)
    resp, err := nc.Request(subject, command, 5*time.Second)
    if err != nil {
        return nil, err
    }
    return resp.Data, nil
}
// 不需要预先订阅，Request 自动创建临时订阅

// 方式 2：通配符订阅响应（如果需要异步处理）
nc.Subscribe("device.*.response", func(msg *nats.Msg) {
    parts := strings.Split(msg.Subject, ".")
    deviceID := parts[1]
    handleResponse(deviceID, msg.Data)
})
// 只需 1 个订阅，覆盖所有设备
```

### 5.4 遥测数据收集

```go
// 收集所有设备的遥测数据

// ✅ 正确：通配符订阅
nc.Subscribe("device.*.telemetry", func(msg *nats.Msg) {
    parts := strings.Split(msg.Subject, ".")
    deviceID := parts[1]
    
    // 存储或处理遥测数据
    storeTelemetry(deviceID, msg.Data)
})

// 如果数据量大，使用 Queue Group 分散到多个实例
nc.QueueSubscribe("device.*.telemetry", "telemetry-processors", handler)
```

---

## 六、JetStream 持久化考虑

### 6.1 Stream 配置

```go
// 百万设备的命令持久化
js.CreateStream(ctx, jetstream.StreamConfig{
    Name:     "DEVICE_COMMANDS",
    Subjects: []string{"device.*.command"},  // 通配符绑定
    MaxAge:   7 * 24 * time.Hour,
    MaxBytes: 100 * 1024 * 1024 * 1024,  // 100GB
})

// 注意：Stream 绑定的是 Subject 模式，不是具体设备 ID
// 百万设备 → 仍然只有 1 个 Stream
```

### 6.2 Consumer 配置

```go
// 按设备 ID 过滤的 Consumer
js.CreateConsumer(ctx, "DEVICE_COMMANDS", jetstream.ConsumerConfig{
    Name:          "device-001-consumer",
    FilterSubject: "device.device-001.command",  // 只消费特定设备
})

// 问题：百万设备 = 百万 Consumer？
// 解决：按需创建，或使用通配符 Consumer
```

---

## 七、总结

### 7.1 核心原则

```
1. Subject 数量不是问题（Subject 只是字符串）
2. 订阅数量才是问题（每个订阅占用内存）
3. 使用通配符订阅压缩订阅数量
4. 设备端各自订阅自己的 Subject（分布式，没问题）
5. 服务端使用通配符或 Request/Reply（避免百万订阅）
```

### 7.2 架构建议

```
设备端（百万进程）：
  每个设备 1 个连接 + 1 个订阅 → ✅ 没问题

Hub 服务端（少数进程）：
  使用通配符订阅 → ✅ 1 个订阅覆盖百万设备
  使用 Request/Reply → ✅ 无需预先订阅

JetStream：
  Stream 绑定 Subject 模式 → ✅ 1 个 Stream
  Consumer 按需创建 → ⚠️ 避免百万 Consumer
```

### 7.3 性能预估

```
百万设备场景：

NATS Server：
  - 连接数：1,000,000（单节点可支持）
  - 内存：约 2-4GB（连接 + 缓冲）
  - 订阅数：约 1,000,000（分布在百万进程中）

Hub 服务：
  - 订阅数：1-10（通配符订阅）
  - 内存：约 100MB

结论：架构合理，性能可行
```