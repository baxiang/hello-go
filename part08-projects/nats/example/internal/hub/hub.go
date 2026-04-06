// Package hub 实现 Hub 中继服务
// 对应项目中的 RelayUseCase + ExecuteService
package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"nats-mvp-example/internal/config"
	"nats-mvp-example/internal/protocol"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Hub 中继服务
// 对应项目中的 RelayUseCase
type Hub struct {
	nc *nats.Conn
	js jetstream.JetStream
	kv jetstream.KeyValue // 设备状态 KV

	// 配置
	cfg *config.HubConfig

	// 响应订阅
	respSub *nats.Subscription

	ctx    context.Context
	cancel context.CancelFunc
}

// New 创建 Hub 实例
func New(cfg *config.HubConfig) (*Hub, error) {
	// 连接 NATS
	nc, err := nats.Connect(cfg.NATSURL,
		nats.Name("hub-server"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(10),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 NATS 失败: %w", err)
	}

	// 创建 JetStream 上下文
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("创建 JetStream 失败: %w", err)
	}

	// 获取 KV Store
	ctx := context.Background()
	kv, err := js.KeyValue(ctx, "DEVICE_STATUS")
	if err != nil {
		// 如果不存在则创建
		kv, err = js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket:       "DEVICE_STATUS",
			Description:  "设备在线状态",
			MaxValueSize: 1024,
			History:      1,
			TTL:          5 * time.Minute,
			Storage:      jetstream.MemoryStorage,
		})
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("创建 KV Store 失败: %w", err)
		}
	}

	h := &Hub{
		nc:  nc,
		js:  js,
		kv:  kv,
		cfg: cfg,
	}

	h.ctx, h.cancel = context.WithCancel(context.Background())

	return h, nil
}

// Start 启动 Hub
func (h *Hub) Start() error {
	// 订阅所有设备的响应
	sub, err := h.nc.Subscribe(protocol.SubjectAllResponse, h.handleResponse)
	if err != nil {
		return fmt.Errorf("订阅响应失败: %w", err)
	}
	h.respSub = sub

	log.Printf("[hub] 已启动，监听设备响应")

	return nil
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	h.cancel()
	if h.respSub != nil {
		h.respSub.Unsubscribe()
	}
	h.nc.Drain()
	h.nc.Close()
}

// IsDeviceOnline 检查设备是否在线
// 对应项目中的 GetDevice
func (h *Hub) IsDeviceOnline(ctx context.Context, deviceID string) bool {
	entry, err := h.kv.Get(ctx, deviceID)
	if err != nil {
		return false
	}

	var status protocol.DeviceStatus
	if err := json.Unmarshal(entry.Value(), &status); err != nil {
		return false
	}

	return status.IsOnline(int64(h.cfg.HeartbeatTimeout.Seconds()))
}

// GetDeviceStatus 获取设备状态
func (h *Hub) GetDeviceStatus(ctx context.Context, deviceID string) (*protocol.DeviceStatus, error) {
	entry, err := h.kv.Get(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("设备 %s 不存在: %w", deviceID, err)
	}

	var status protocol.DeviceStatus
	if err := json.Unmarshal(entry.Value(), &status); err != nil {
		return nil, fmt.Errorf("解析状态失败: %w", err)
	}

	return &status, nil
}

// ExecuteResult 执行结果
type ExecuteResult struct {
	MsgID     string          // 消息 ID
	Type      string          // 响应类型
	Seq       int             // 序列号
	Data      json.RawMessage // 响应数据
	Error     string          // 错误信息
	Timestamp int64           // 时间戳
}

// ExecuteCommand 执行命令
// 对应项目中的 SendQuery + SSE 流式响应
// 使用 Request/Reply 模式实现流式响应
func (h *Hub) ExecuteCommand(ctx context.Context, deviceID, query string) (<-chan *ExecuteResult, error) {
	// 1. 检查设备在线
	if !h.IsDeviceOnline(ctx, deviceID) {
		return nil, fmt.Errorf("设备 %s 不在线", deviceID)
	}

	// 2. 创建请求
	req := protocol.NewCommandRequest(query)
	reqData, _ := json.Marshal(req)

	// 3. 创建响应订阅（使用 Inbox）
	inbox := nats.NewInbox()
	respCh := make(chan *ExecuteResult, 10)

	sub, err := h.nc.Subscribe(inbox, func(msg *nats.Msg) {
		var resp protocol.Response
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			log.Printf("[hub] 解析响应失败: %v", err)
			return
		}

		result := &ExecuteResult{
			MsgID:     resp.MsgID,
			Type:      resp.Type,
			Seq:       resp.Seq,
			Data:      resp.Data,
			Error:     resp.Error,
			Timestamp: resp.Timestamp,
		}

		select {
		case respCh <- result:
		case <-ctx.Done():
		}

		// 如果是完成或错误，关闭通道
		if resp.Type == protocol.TypeResponseDone || resp.Type == protocol.TypeError {
			close(respCh)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("创建响应订阅失败: %w", err)
	}

	// 4. 发送命令
	commandSubject := protocol.GetCommandSubject(deviceID)
	msg := &nats.Msg{
		Subject: commandSubject,
		Reply:   inbox,
		Data:    reqData,
	}

	if err := h.nc.PublishMsg(msg); err != nil {
		sub.Unsubscribe()
		close(respCh)
		return nil, fmt.Errorf("发送命令失败: %w", err)
	}

	log.Printf("[hub] 发送命令: device=%s, msg_id=%s", deviceID, req.MsgID)

	// 5. 启动超时清理
	go func() {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
		case <-time.After(h.cfg.ExecuteTimeout):
			sub.Unsubscribe()
			close(respCh)
		}
	}()

	return respCh, nil
}

// ExecuteCommandSync 同步执行命令（等待完成）
func (h *Hub) ExecuteCommandSync(ctx context.Context, deviceID, query string) (*ExecuteResult, error) {
	respCh, err := h.ExecuteCommand(ctx, deviceID, query)
	if err != nil {
		return nil, err
	}

	var lastResult *ExecuteResult
	for result := range respCh {
		if result.Type == protocol.TypeResponseDone || result.Type == protocol.TypeError {
			lastResult = result
			break
		}
		lastResult = result
	}

	return lastResult, nil
}

// CancelCommand 取消命令
// 对应项目中的 cancel 消息
func (h *Hub) CancelCommand(ctx context.Context, deviceID, msgID string) error {
	req := protocol.NewCancelRequest(msgID)
	data, _ := json.Marshal(req)

	subject := protocol.GetCancelSubject(deviceID)
	return h.nc.Publish(subject, data)
}

// handleResponse 处理设备响应（通配符订阅）
func (h *Hub) handleResponse(msg *nats.Msg) {
	// 从 Subject 解析设备 ID
	// device.{device_id}.response
	// 这里只是日志记录，实际响应通过 Reply-To 处理
	log.Printf("[hub] 收到响应: subject=%s, len=%d", msg.Subject, len(msg.Data))
}

// WatchDeviceStatus 监听设备状态变更
// 对应项目中的设备上线/下线通知
func (h *Hub) WatchDeviceStatus(ctx context.Context) (<-chan *protocol.DeviceStatus, error) {
	watcher, err := h.kv.WatchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建状态监听失败: %w", err)
	}

	statusCh := make(chan *protocol.DeviceStatus, 10)

	go func() {
		defer close(statusCh)
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				var status protocol.DeviceStatus
				if err := json.Unmarshal(entry.Value(), &status); err != nil {
					continue
				}

				statusCh <- &status
			}
		}
	}()

	return statusCh, nil
}

// ListOnlineDevices 列出所有在线设备
func (h *Hub) ListOnlineDevices(ctx context.Context) ([]*protocol.DeviceStatus, error) {
	keys, err := h.kv.Keys(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取设备列表失败: %w", err)
	}

	var devices []*protocol.DeviceStatus
	for _, key := range keys {
		entry, err := h.kv.Get(ctx, key)
		if err != nil {
			continue
		}

		var status protocol.DeviceStatus
		if err := json.Unmarshal(entry.Value(), &status); err != nil {
			continue
		}

		if status.IsOnline(int64(h.cfg.HeartbeatTimeout.Seconds())) {
			devices = append(devices, &status)
		}
	}

	return devices, nil
}

// PendingRequests 获取待处理请求数量
// 对应项目中的 PendingCount
// 注意：NATS 版本不需要维护 Pending Map，由 Inbox 自动管理
func (h *Hub) PendingRequests() int {
	// NATS 版本不需要手动管理 Pending
	// 每个 Request 使用独立的 Inbox
	return 0
}

// HubStats Hub 统计信息
type HubStats struct {
	TotalDevices    int // 总设备数
	OnlineDevices   int // 在线设备数
	PendingRequests int // 待处理请求数
}

// GetStats 获取统计信息
func (h *Hub) GetStats(ctx context.Context) (*HubStats, error) {
	devices, err := h.ListOnlineDevices(ctx)
	if err != nil {
		return nil, err
	}

	return &HubStats{
		OnlineDevices:   len(devices),
		PendingRequests: h.PendingRequests(),
	}, nil
}

// DevicePool 设备连接池
// 用于管理多个设备连接（测试用）
type DevicePool struct {
	hub     *Hub
	devices sync.Map // deviceID → *DeviceConnection
}

// NewDevicePool 创建设备连接池
func NewDevicePool(hub *Hub) *DevicePool {
	return &DevicePool{
		hub: hub,
	}
}

// GetDevice 获取设备连接
// 对应项目中的 sync.Map clients
func (p *DevicePool) GetDevice(deviceID string) (*DeviceConnection, bool) {
	v, ok := p.devices.Load(deviceID)
	if !ok {
		return nil, false
	}
	return v.(*DeviceConnection), true
}

// DeviceConnection 设备连接状态
// 对应项目中的 ClientConnection
type DeviceConnection struct {
	DeviceID  string
	AccountID string
	NodeType  string
	Status    string
}