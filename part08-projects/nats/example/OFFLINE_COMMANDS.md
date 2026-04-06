# 设备离线消息处理方案

> 解决设备下线/销毁时消息丢失问题

## 一、问题分析

### 1.1 场景描述

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     设备下线/销毁场景                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  场景 1：设备离线时发送命令                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Client ──Request──> Hub ──Publish──> NATS ──X──> Device (离线)     │   │
│  │                                                                     │   │
│  │  结果：消息丢失，客户端超时                                          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  场景 2：设备处理中突然下线                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Device 收到命令 ──处理中──> 突然断电/崩溃                           │   │
│  │                                                                     │   │
│  │  结果：处理中的消息丢失，客户端超时                                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  场景 3：设备销毁（永久删除）                                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  设备被删除，但还有未处理的命令                                       │   │
│  │                                                                     │   │
│  │  结果：消息永久丢失                                                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Core NATS 的局限性

```
Core NATS Pub/Sub 特点：

  ✅ 优点：
    - 性能极高（18M msg/s）
    - 延迟低（< 1ms）
    - 实现简单

  ❌ 局限性：
    - At Most Once（最多一次投递）
    - 无持久化，订阅者不在线则消息丢弃
    - 发布者无法感知消息是否被消费
    - 无重试机制
```

## 二、解决方案

### 2.1 方案对比

| 方案 | 消息不丢失 | 实时性 | 复杂度 | 适用场景 |
|------|-----------|--------|--------|----------|
| 纯 Request/Reply | ❌ | 高 | 低 | 设备始终在线 |
| JetStream 持久化 | ✅ | 中 | 中 | 需要可靠投递 |
| 混合模式 | ✅ | 高 | 高 | 生产环境推荐 |
| 命令队列 + 状态机 | ✅ | 中 | 高 | 需要命令追踪 |

### 2.2 推荐方案：混合模式

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     混合模式架构                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  设备在线                                                            │   │
│  │                                                                     │   │
│  │  Hub ──Request/Reply──> NATS ──Subscribe──> Device                  │   │
│  │                                                                     │   │
│  │  特点：                                                              │   │
│  │    - 实时性高（< 1ms）                                               │   │
│  │    - 无持久化开销                                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  设备离线                                                            │   │
│  │                                                                     │   │
│  │  Hub ──Publish──> JetStream Stream ──Consume──> Device (上线后)     │   │
│  │                                                                     │   │
│  │  特点：                                                              │   │
│  │    - 消息不丢失                                                      │   │
│  │    - 设备上线后继续处理                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 三、架构设计

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     NATS 混合模式完整架构                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Client                    Hub                    NATS                      │
│     │                       │                       │                       │
│     │  1. 发送命令          │                       │                       │
│     │──────────────────────>│                       │                       │
│     │                       │                       │                       │
│     │                       │  2. 检查设备状态       │                       │
│     │                       │  KV.Get(deviceID)     │                       │
│     │                       │──────────────────────>│                       │
│     │                       │<──────────────────────│                       │
│     │                       │                       │                       │
│     │                       │  ┌─────────────────┐  │                       │
│     │                       │  │ 设备在线？      │  │                       │
│     │                       │  └────────┬────────┘  │                       │
│     │                       │           │           │                       │
│     │                       │     ┌─────┴─────┐     │                       │
│     │                       │     │           │     │                       │
│     │                       │   在线         离线   │                       │
│     │                       │     │           │     │                       │
│     │                       │     ▼           ▼     │                       │
│     │                       │                       │                       │
│     │                       │  Request/Reply  Publish to Stream             │
│     │                       │───────────────────────────────────────────────>│
│     │                       │                       │                       │
│     │                       │                       │  Stream 持久化        │
│     │                       │                       │  ┌─────────────────┐  │
│     │                       │                       │  │ DEVICE_001      │  │
│     │                       │                       │  │ - msg1 (pending)│  │
│     │                       │                       │  │ - msg2 (pending)│  │
│     │                       │                       │  └─────────────────┘  │
│     │                       │                       │                       │
│                                                                             │
│  Device (上线后)                                                            │
│     │                       │                       │                       │
│     │  3. 注册状态          │                       │                       │
│     │───────────────────────────────────────────────>│                       │
│     │                       │                       │                       │
│     │  4. 订阅实时命令      │                       │                       │
│     │<──────────────────────────────────────────────│                       │
│     │                       │                       │                       │
│     │  5. 消费离线命令      │                       │                       │
│     │<──────────────────────────────────────────────│                       │
│     │                       │                       │                       │
│     │  6. 处理并响应        │                       │                       │
│     │──────────────────────>│                       │                       │
│     │                       │                       │                       │
│     │  7. 推送结果给 Client │                       │                       │
│     │                       │<──────────────────────│                       │
│     │<──────────────────────│                       │                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Subject 设计

```
实时命令（Core NATS）：
  device.{device_id}.command      # 设备订阅，接收实时命令
  device.{device_id}.cancel       # 取消正在执行的命令

离线命令（JetStream Stream）：
  Stream: DEVICE_{device_id}_COMMANDS
  Subject: device.{device_id}.command  # 与实时命令相同 Subject

设备状态（KV Store）：
  Bucket: DEVICE_STATUS
  Key: {device_id}

命令状态（KV Store）：
  Bucket: COMMAND_STATUS
  Key: {command_id}
```

### 3.3 命令状态机

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     命令状态机                                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│  │ PENDING │───>│ SENT    │───>│ RUNNING │───>│ DONE    │───>│ ARCHIVED│  │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘  │
│       │              │              │              │                       │
│       │              │              │              │                       │
│       ▼              ▼              ▼              ▼                       │
│  设备离线        已发送          处理中          已完成                    │
│  等待上线        等待响应        流式响应        可归档                    │
│       │              │              │                                      │
│       │              │              │                                      │
│       ▼              ▼              ▼                                      │
│  超时失败        超时重试        设备离线                                   │
│                               暂停等待                                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 四、代码实现

### 4.1 消息协议

```go
// internal/protocol/command.go

package protocol

import "time"

// CommandState 命令状态
type CommandState string

const (
    CommandPending  CommandState = "pending"   // 等待发送
    CommandSent     CommandState = "sent"      // 已发送，等待响应
    CommandRunning  CommandState = "running"   // 处理中
    CommandDone     CommandState = "done"      // 已完成
    CommandFailed   CommandState = "failed"    // 失败
    CommandCanceled CommandState = "canceled"  // 已取消
)

// CommandRecord 命令记录（存储在 KV Store）
type CommandRecord struct {
    ID           string        `json:"id"`
    DeviceID     string        `json:"device_id"`
    Query        string        `json:"query"`
    State        CommandState  `json:"state"`
    CreatedAt    int64         `json:"created_at"`
    UpdatedAt    int64         `json:"updated_at"`
    Result       string        `json:"result,omitempty"`
    Error        string        `json:"error,omitempty"`
    RetryCount   int           `json:"retry_count"`
    MaxRetry     int           `json:"max_retry"`
    ExpireAt     int64         `json:"expire_at"`     // 命令过期时间
    ResponseChan string        `json:"response_chan"` // 响应通道（用于推送结果）
}

// NewCommandRecord 创建命令记录
func NewCommandRecord(deviceID, query string) *CommandRecord {
    now := time.Now().Unix()
    return &CommandRecord{
        ID:         generateMsgID(),
        DeviceID:   deviceID,
        Query:      query,
        State:      CommandPending,
        CreatedAt:  now,
        UpdatedAt:  now,
        MaxRetry:   3,
        ExpireAt:   now + 86400, // 24 小时过期
    }
}

// CanRetry 是否可以重试
func (c *CommandRecord) CanRetry() bool {
    return c.RetryCount < c.MaxRetry
}

// IsExpired 是否已过期
func (c *CommandRecord) IsExpired() bool {
    return time.Now().Unix() > c.ExpireAt
}
```

### 4.2 Hub 端实现

```go
// internal/hub/command.go

package hub

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "nats-mvp-example/internal/protocol"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

// CommandManager 命令管理器
type CommandManager struct {
    nc *nats.Conn
    js jetstream.JetStream
    
    // KV Store
    deviceKV  jetstream.KeyValue  // 设备状态
    commandKV jetstream.KeyValue  // 命令状态
    
    // 待处理响应
    pending sync.Map  // commandID → chan *protocol.Response
    
    // 配置
    cfg *CommandConfig
}

// CommandConfig 命令配置
type CommandConfig struct {
    DefaultTimeout   time.Duration  // 默认超时
    MaxRetry         int            // 最大重试次数
    CommandExpire    time.Duration  // 命令过期时间
    OfflineQueueSize int            // 离线队列大小
}

// ExecuteCommand 执行命令
func (m *CommandManager) ExecuteCommand(ctx context.Context, deviceID, query string) (<-chan *protocol.Response, error) {
    // 1. 创建命令记录
    cmd := protocol.NewCommandRecord(deviceID, query)
    cmd.MaxRetry = m.cfg.MaxRetry
    cmd.ExpireAt = time.Now().Add(m.cfg.CommandExpire).Unix()
    
    // 2. 存储命令状态
    cmdKey := fmt.Sprintf("command.%s", cmd.ID)
    cmdData, _ := json.Marshal(cmd)
    m.commandKV.Put(ctx, cmdKey, cmdData)
    
    // 3. 创建响应通道
    respCh := make(chan *protocol.Response, 10)
    m.pending.Store(cmd.ID, respCh)
    
    // 4. 检查设备状态并发送
    go m.processCommand(ctx, cmd)
    
    return respCh, nil
}

// processCommand 处理命令
func (m *CommandManager) processCommand(ctx context.Context, cmd *protocol.CommandRecord) {
    // 检查设备是否在线
    online := m.isDeviceOnline(ctx, cmd.DeviceID)
    
    if online {
        // 设备在线：直接发送
        m.sendToOnlineDevice(ctx, cmd)
    } else {
        // 设备离线：存入 Stream
        m.sendToOfflineQueue(ctx, cmd)
    }
}

// isDeviceOnline 检查设备是否在线
func (m *CommandManager) isDeviceOnline(ctx context.Context, deviceID string) bool {
    entry, err := m.deviceKV.Get(ctx, deviceID)
    if err != nil {
        return false
    }
    
    var status protocol.DeviceStatus
    json.Unmarshal(entry.Value(), &status)
    
    return status.Status == "online"
}

// sendToOnlineDevice 发送给在线设备
func (m *CommandManager) sendToOnlineDevice(ctx context.Context, cmd *protocol.CommandRecord) {
    // 更新状态为 sent
    m.updateCommandState(ctx, cmd, protocol.CommandSent)
    
    // 创建响应订阅
    inbox := nats.NewInbox()
    sub, err := m.nc.Subscribe(inbox, func(msg *nats.Msg) {
        var resp protocol.Response
        json.Unmarshal(msg.Data, &resp)
        
        // 推送响应
        if ch, ok := m.pending.Load(cmd.ID); ok {
            ch.(chan *protocol.Response) <- &resp
        }
        
        // 如果是完成或错误，更新状态并清理
        if resp.Type == protocol.TypeResponseDone || resp.Type == protocol.TypeError {
            m.updateCommandState(ctx, cmd, protocol.CommandDone)
            m.pending.Delete(cmd.ID)
            sub.Unsubscribe()
        }
    })
    if err != nil {
        m.handleSendError(ctx, cmd, err)
        return
    }
    
    // 发送命令
    subject := fmt.Sprintf("device.%s.command", cmd.DeviceID)
    reqData, _ := json.Marshal(&protocol.CommandRequest{
        MsgID:     cmd.ID,
        Query:     cmd.Query,
        Timestamp: time.Now().Unix(),
    })
    
    err = m.nc.PublishMsg(&nats.Msg{
        Subject: subject,
        Reply:   inbox,
        Data:    reqData,
    })
    
    if err != nil {
        sub.Unsubscribe()
        m.handleSendError(ctx, cmd, err)
        return
    }
    
    // 设置超时
    go m.watchTimeout(ctx, cmd, sub)
}

// sendToOfflineQueue 发送到离线队列
func (m *CommandManager) sendToOfflineQueue(ctx context.Context, cmd *protocol.CommandRecord) {
    // 获取或创建设备专属 Stream
    streamName := fmt.Sprintf("DEVICE_%s_COMMANDS", cmd.DeviceID)
    
    _, err := m.js.CreateStream(ctx, jetstream.StreamConfig{
        Name:       streamName,
        Subjects:   []string{fmt.Sprintf("device.%s.command", cmd.DeviceID)},
        MaxAge:     m.cfg.CommandExpire,
        MaxMsgs:    int64(m.cfg.OfflineQueueSize),
        Storage:    jetstream.MemoryStorage,
    })
    if err != nil {
        log.Printf("[hub] 创建 Stream 失败: %v", err)
    }
    
    // 发布到 Stream
    reqData, _ := json.Marshal(&protocol.CommandRequest{
        MsgID:     cmd.ID,
        Query:     cmd.Query,
        Timestamp: time.Now().Unix(),
    })
    
    subject := fmt.Sprintf("device.%s.command", cmd.DeviceID)
    _, err = m.js.Publish(ctx, subject, reqData)
    if err != nil {
        m.handleSendError(ctx, cmd, err)
        return
    }
    
    log.Printf("[hub] 命令已存入离线队列: device=%s, cmd_id=%s", cmd.DeviceID, cmd.ID)
}

// handleSendError 处理发送错误
func (m *CommandManager) handleSendError(ctx context.Context, cmd *protocol.CommandRecord, err error) {
    log.Printf("[hub] 发送命令失败: %v", err)
    
    if cmd.CanRetry() {
        // 重试
        cmd.RetryCount++
        m.updateCommandState(ctx, cmd, protocol.CommandPending)
        time.Sleep(time.Second * time.Duration(cmd.RetryCount))
        go m.processCommand(ctx, cmd)
    } else {
        // 失败
        cmd.Error = err.Error()
        m.updateCommandState(ctx, cmd, protocol.CommandFailed)
        
        // 通知客户端
        if ch, ok := m.pending.Load(cmd.ID); ok {
            ch.(chan *protocol.Response) <- &protocol.Response{
                MsgID: cmd.ID,
                Type:  protocol.TypeError,
                Error: err.Error(),
            }
            m.pending.Delete(cmd.ID)
        }
    }
}

// watchTimeout 监控超时
func (m *CommandManager) watchTimeout(ctx context.Context, cmd *protocol.CommandRecord, sub *nats.Subscription) {
    time.Sleep(m.cfg.DefaultTimeout)
    
    // 检查命令是否已完成
    cmdKey := fmt.Sprintf("command.%s", cmd.ID)
    entry, err := m.commandKV.Get(ctx, cmdKey)
    if err != nil {
        return
    }
    
    var currentCmd protocol.CommandRecord
    json.Unmarshal(entry.Value(), &currentCmd)
    
    if currentCmd.State == protocol.CommandSent || currentCmd.State == protocol.CommandRunning {
        // 超时
        sub.Unsubscribe()
        m.handleSendError(ctx, cmd, fmt.Errorf("命令超时"))
    }
}

// updateCommandState 更新命令状态
func (m *CommandManager) updateCommandState(ctx context.Context, cmd *protocol.CommandRecord, state protocol.CommandState) {
    cmd.State = state
    cmd.UpdatedAt = time.Now().Unix()
    
    cmdKey := fmt.Sprintf("command.%s", cmd.ID)
    cmdData, _ := json.Marshal(cmd)
    m.commandKV.Put(ctx, cmdKey, cmdData)
}

// OnDeviceOnline 设备上线回调
func (m *CommandManager) OnDeviceOnline(ctx context.Context, deviceID string) {
    log.Printf("[hub] 设备上线，检查离线命令: device=%s", deviceID)
    
    // 查询该设备的 pending 命令
    // 这里可以通过 Stream Consumer 来消费积压的消息
    streamName := fmt.Sprintf("DEVICE_%s_COMMANDS", deviceID)
    
    consumer, err := m.js.CreateConsumer(ctx, streamName, jetstream.ConsumerConfig{
        Name:     "hub-processor",
        AckPolicy: jetstream.AckExplicit,
    })
    if err != nil {
        return
    }
    
    // 消费积压消息
    msgs, _ := consumer.Fetch(100)
    for msg := range msgs {
        var req protocol.CommandRequest
        json.Unmarshal(msg.Data(), &req)
        
        // 重新处理命令
        cmdKey := fmt.Sprintf("command.%s", req.MsgID)
        entry, err := m.commandKV.Get(ctx, cmdKey)
        if err != nil {
            msg.Ack()
            continue
        }
        
        var cmd protocol.CommandRecord
        json.Unmarshal(entry.Value(), &cmd)
        
        // 发送给在线设备
        go m.sendToOnlineDevice(ctx, &cmd)
        
        // 确认消息
        msg.Ack()
    }
}
```

### 4.3 Device 端实现

```go
// internal/device/command.go

package device

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "nats-mvp-example/internal/protocol"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

// Device 设备端
type Device struct {
    id   string
    nc   *nats.Conn
    js   jetstream.JetStream
    kv   jetstream.KeyValue
    
    ctx    context.Context
    cancel context.CancelFunc
}

// Start 启动设备
func (d *Device) Start(ctx context.Context) error {
    // 1. 注册设备状态
    d.registerStatus(ctx)
    
    // 2. 订阅实时命令
    d.subscribeRealtimeCommands(ctx)
    
    // 3. 消费离线命令
    d.consumeOfflineCommands(ctx)
    
    // 4. 启动心跳
    go d.heartbeatLoop(ctx)
    
    log.Printf("[device-%s] 已启动", d.id)
    
    <-ctx.Done()
    return nil
}

// subscribeRealtimeCommands 订阅实时命令
func (d *Device) subscribeRealtimeCommands(ctx context.Context) {
    subject := fmt.Sprintf("device.%s.command", d.id)
    
    d.nc.Subscribe(subject, func(msg *nats.Msg) {
        var req protocol.CommandRequest
        if err := json.Unmarshal(msg.Data, &req); err != nil {
            log.Printf("[device-%s] 解析命令失败: %v", d.id, err)
            return
        }
        
        log.Printf("[device-%s] 收到实时命令: msg_id=%s", d.id, req.MsgID)
        
        // 处理命令
        d.handleCommand(ctx, msg.Reply, &req)
    })
}

// consumeOfflineCommands 消费离线命令
func (d *Device) consumeOfflineCommands(ctx context.Context) {
    streamName := fmt.Sprintf("DEVICE_%s_COMMANDS", d.id)
    
    // 创建 Consumer
    consumer, err := d.js.CreateConsumer(ctx, streamName, jetstream.ConsumerConfig{
        Name:      "device-consumer",
        Durable:   "device-consumer",
        AckPolicy: jetstream.AckExplicit,
    })
    if err != nil {
        log.Printf("[device-%s] 无离线命令", d.id)
        return
    }
    
    // 消费消息
    consumer.Consume(func(msg jetstream.Msg) {
        var req protocol.CommandRequest
        if err := json.Unmarshal(msg.Data(), &req); err != nil {
            msg.Ack()
            return
        }
        
        log.Printf("[device-%s] 收到离线命令: msg_id=%s", d.id, req.MsgID)
        
        // 处理命令（使用临时 Reply）
        inbox := nats.NewInbox()
        d.handleCommand(ctx, inbox, &req)
        
        // 确认处理完成
        msg.Ack()
    })
}

// handleCommand 处理命令
func (d *Device) handleCommand(ctx context.Context, replyTo string, req *protocol.CommandRequest) {
    // 模拟处理
    chunks := []string{
        "正在处理...\n",
        "分析中...\n",
        "完成。",
    }
    
    // 发送流式响应
    for i, chunk := range chunks {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        resp := protocol.NewChunkResponse(req.MsgID, i+1, []byte(chunk))
        data, _ := json.Marshal(resp)
        d.nc.Publish(replyTo, data)
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // 发送完成响应
    doneResp := protocol.NewDoneResponse(req.MsgID)
    data, _ := json.Marshal(doneResp)
    d.nc.Publish(replyTo, data)
    
    log.Printf("[device-%s] 命令完成: msg_id=%s", d.id, req.MsgID)
}

// registerStatus 注册设备状态
func (d *Device) registerStatus(ctx context.Context) {
    status := protocol.NewDeviceStatus(d.id, "account-001", "agent")
    data, _ := json.Marshal(status)
    d.kv.Put(ctx, d.id, data)
}

// heartbeatLoop 心跳循环
func (d *Device) heartbeatLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            d.updateHeartbeat(ctx)
        }
    }
}

// updateHeartbeat 更新心跳
func (d *Device) updateHeartbeat(ctx context.Context) {
    entry, err := d.kv.Get(ctx, d.id)
    if err != nil {
        d.registerStatus(ctx)
        return
    }
    
    var status protocol.DeviceStatus
    json.Unmarshal(entry.Value(), &status)
    status.UpdateHeartbeat()
    
    data, _ := json.Marshal(status)
    d.kv.Put(ctx, d.id, data)
}
```

## 五、配置说明

```go
// internal/config/command.go

package config

import "time"

// CommandConfig 命令配置
type CommandConfig struct {
    // 超时设置
    DefaultTimeout   time.Duration `yaml:"default_timeout"`    // 默认超时（60s）
    FirstChunkTimeout time.Duration `yaml:"first_chunk_timeout"` // 首字超时（10s）
    
    // 重试设置
    MaxRetry      int           `yaml:"max_retry"`       // 最大重试次数（3）
    RetryInterval time.Duration `yaml:"retry_interval"`  // 重试间隔（1s）
    
    // 过期设置
    CommandExpire time.Duration `yaml:"command_expire"` // 命令过期时间（24h）
    
    // 离线队列
    OfflineQueueSize int `yaml:"offline_queue_size"` // 离线队列大小（1000）
    
    // 流式响应
    ResponseBufferSize int `yaml:"response_buffer_size"` // 响应缓冲区大小（10）
}

// DefaultCommandConfig 默认配置
func DefaultCommandConfig() *CommandConfig {
    return &CommandConfig{
        DefaultTimeout:      60 * time.Second,
        FirstChunkTimeout:   10 * time.Second,
        MaxRetry:            3,
        RetryInterval:       1 * time.Second,
        CommandExpire:       24 * time.Hour,
        OfflineQueueSize:    1000,
        ResponseBufferSize:  10,
    }
}
```

## 六、使用示例

### 6.1 Hub 端

```go
// 发送命令
respCh, err := hub.ExecuteCommand(ctx, "device-001", "你好")
if err != nil {
    log.Fatal(err)
}

// 接收流式响应
for resp := range respCh {
    if resp.Type == "response_chunk" {
        fmt.Print(string(resp.Data))
    } else if resp.Type == "response_done" {
        fmt.Println("\n完成")
        break
    } else if resp.Type == "error" {
        fmt.Println("错误:", resp.Error)
        break
    }
}
```

### 6.2 Device 端

```go
// 启动设备
device := NewDevice(cfg)
device.Start(ctx)
```

## 七、监控指标

```
命令状态分布：
  - pending: 等待发送的命令数
  - sent: 已发送等待响应的命令数
  - running: 正在处理的命令数
  - done: 已完成的命令数
  - failed: 失败的命令数

设备状态：
  - online_devices: 在线设备数
  - offline_devices: 离线设备数

离线队列：
  - offline_queue_size: 离线队列大小
  - offline_queue_messages: 离线消息数

性能指标：
  - command_latency: 命令延迟
  - command_success_rate: 命令成功率
  - retry_count: 重试次数
```

## 八、总结

| 场景 | 解决方案 | 说明 |
|------|----------|------|
| 设备在线 | Request/Reply | 实时性高 |
| 设备离线 | JetStream Stream | 消息持久化 |
| 设备上线 | 消费离线命令 | 继续处理 |
| 命令超时 | 重试机制 | 可配置重试次数 |
| 命令过期 | TTL | 自动清理过期命令 |
| 命令追踪 | KV Store | 状态可查询 |