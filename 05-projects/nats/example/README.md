# NATS 架构迁移示例

> 基于 livis-claw-hub 项目实际架构，验证 NATS 替代 WebSocket 的可行性

## 一、项目结构

```
example/
├── cmd/
│   ├── device/main.go      # 设备端程序
│   ├── hub/main.go         # Hub 服务程序
│   ├── client/main.go      # HTTP 客户端
│   └── test/main.go        # 端到端测试
├── internal/
│   ├── protocol/           # 消息协议（对应项目 WebSocket 协议）
│   ├── device/             # 设备端实现（对应项目 ClientConnection）
│   ├── hub/                # Hub 服务实现（对应项目 RelayUseCase）
│   ├── client/             # 客户端实现
│   └── config/             # 配置
├── go.mod
└── README.md
```

## 二、架构对比

### 2.1 组件映射

| 当前组件 | NATS 替代 | 文件 |
|----------|-----------|------|
| WebSocket.Conn | nats.Conn | device/device.go |
| sync.Map clients | KV Store | hub/hub.go |
| map[string]chan Pending | Request/Reply + Inbox | hub/hub.go |
| SSE 流式响应 | 多消息响应 | hub/hub.go |
| Ping/Pong 心跳 | KV 更新 | device/device.go |

### 2.2 Subject 设计

```
设备端订阅：
  device.{device_id}.command    # 接收命令
  device.{device_id}.cancel     # 接收取消

设备端发布：
  device.{device_id}.response   # 响应（通过 Reply-To）

Hub 订阅：
  device.*.response             # 通配符订阅所有响应

设备状态（KV Store）：
  Bucket: DEVICE_STATUS
  Key: {device_id}
```

### 2.3 消息协议映射

**WebSocket 协议：**
```json
{"type":"query","msg_id":"xxx","payload":{"query":"..."}}
{"type":"response_chunk","msg_id":"xxx","payload":"..."}
{"type":"response_done","msg_id":"xxx","payload":"..."}
```

**NATS 协议：**
```json
// 请求
{"msg_id":"xxx","query":"...","timestamp":1234567890}

// 流式响应
{"msg_id":"xxx","type":"response_chunk","seq":1,"data":"..."}
{"msg_id":"xxx","type":"response_done","seq":5}
```

## 三、运行示例

### 3.1 端到端测试

```bash
# 运行完整测试
go run ./cmd/test -mode=test

# 输出：
# === 端到端测试 ===
# --- 测试 1: 检查设备在线 ---
# 设备在线: true
# --- 测试 2: 获取设备状态 ---
# 设备状态: id=test-device-001, status=online
# --- 测试 3: 列出在线设备 ---
# 在线设备数: 3
# --- 测试 4: 执行命令 ---
# 响应: type=response_chunk, seq=1, data="正在处理..."
# 响应: type=response_done, seq=5
# === 测试完成 ===
```

### 3.2 独立运行

```bash
# 终端 1：启动 Hub
go run ./cmd/hub -nats=nats://localhost:4222 -http=:8080

# 终端 2：启动设备
go run ./cmd/device -id=device-001 -nats=nats://localhost:4222

# 终端 3：发送命令
curl -X POST "http://localhost:8080/api/v1/device/execute?device_id=device-001" \
  -H "Content-Type: application/json" \
  -d '{"query":"你好"}'
```

## 四、核心实现

### 4.1 设备注册（替代 sync.Map）

```go
// 当前：sync.Map
clients.Store(deviceID, client)

// NATS：KV Store
status := &DeviceStatus{DeviceID: deviceID, Status: "online"}
kv.Put(ctx, deviceID, json.Marshal(status))
```

### 4.2 命令执行（替代 WebSocket + Pending Map）

```go
// 当前：WebSocket + Pending Map
waitCh := make(chan *WSMessage)
client.Pending[msgID] = waitCh
client.WriteJSON(queryMsg)

// NATS：Request/Reply + Inbox
inbox := nats.NewInbox()
nc.Subscribe(inbox, handleResponse)
nc.PublishMsg(&nats.Msg{
    Subject: "device." + deviceID + ".command",
    Reply:   inbox,
    Data:    requestData,
})
```

### 4.3 流式响应

```go
// 设备端：发送多个响应
for i, chunk := range chunks {
    resp := NewChunkResponse(msgID, i+1, chunk)
    nc.Publish(replyTo, json.Marshal(resp))
}
resp := NewDoneResponse(msgID)
nc.Publish(replyTo, json.Marshal(resp))
```

## 五、性能分析

详见 [PERFORMANCE.md](./PERFORMANCE.md)

### 5.1 百万设备场景

| 维度 | 当前架构 | NATS 架构 |
|------|----------|-----------|
| 连接数 | 单节点 10 万级 | 单节点 100 万级 |
| 设备注册 | sync.Map 内存 | KV Store 分布式 |
| 跨节点协调 | 需自研 Redis | 内置 Cluster |
| 延迟 | < 10ms | < 1ms |

### 5.2 Subject 数量

```
百万设备 = 百万 Subject？

✅ 没问题！
- Subject 只是字符串，不占用内存
- 每个设备只创建 1 个订阅
- Hub 使用通配符订阅（1 个订阅覆盖百万设备）
```

## 六、迁移路径

1. **阶段一**：并行运行，验证 NATS 方案
2. **阶段二**：灰度切换，逐步迁移设备
3. **阶段三**：完全迁移，下线 WebSocket

## 七、相关文档

- [PERFORMANCE.md](./PERFORMANCE.md) - 百万设备性能分析
- [OFFLINE_COMMANDS.md](./OFFLINE_COMMANDS.md) - 设备离线消息处理方案
- [../hands-on/](../hands-on/) - NATS 边学边练教程