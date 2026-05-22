// Package device 实现设备端 NATS 连接和消息处理
// 对应项目中的 WebSocket 设备连接
package device

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

// Device 设备端
// 对应项目中的 ClientConnection，但角色是设备端
type Device struct {
	id        string          // 设备 ID
	accountID string          // 账号 ID
	nodeType  string          // 节点类型
	nc        *nats.Conn      // NATS 连接
	js        jetstream.JetStream
	kv        jetstream.KeyValue // 设备状态 KV

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 配置
	cfg *config.DeviceConfig
}

// New 创建设备实例
func New(cfg *config.DeviceConfig) (*Device, error) {
	// 连接 NATS
	nc, err := nats.Connect(cfg.NATSURL,
		nats.Name(fmt.Sprintf("device-%s", cfg.ID)),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("[device-%s] 断开连接: %v", cfg.ID, err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[device-%s] 重新连接: %s", cfg.ID, nc.ConnectedUrl())
		}),
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

	// 获取或创建 KV Store
	ctx := context.Background()
	kv, err := js.KeyValue(ctx, "DEVICE_STATUS")
	if err != nil {
		// 如果不存在则创建
		kv, err = js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket:       "DEVICE_STATUS",
			Description:  "设备在线状态",
			MaxValueSize: 1024,
			History:      1,
			TTL:          5 * time.Minute, // 心跳超时
			Storage:      jetstream.MemoryStorage,
		})
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("创建 KV Store 失败: %w", err)
		}
	}

	d := &Device{
		id:        cfg.ID,
		accountID: cfg.AccountID,
		nodeType:  cfg.NodeType,
		nc:        nc,
		js:        js,
		kv:        kv,
		cfg:       cfg,
	}

	d.ctx, d.cancel = context.WithCancel(context.Background())

	return d, nil
}

// Start 启动设备
// 对应项目中的 WebSocket 连接和消息读取循环
func (d *Device) Start() error {
	// 1. 注册设备状态
	if err := d.registerStatus(); err != nil {
		return err
	}

	// 2. 订阅命令 Subject
	commandSub, err := d.subscribeCommand()
	if err != nil {
		d.unregisterStatus()
		return err
	}

	// 3. 订阅取消 Subject
	cancelSub, err := d.subscribeCancel()
	if err != nil {
		commandSub.Unsubscribe()
		d.unregisterStatus()
		return err
	}

	// 4. 启动心跳
	d.wg.Add(1)
	go d.heartbeatLoop()

	log.Printf("[device-%s] 已启动，监听命令", d.id)

	// 等待退出
	<-d.ctx.Done()

	// 清理
	commandSub.Unsubscribe()
	cancelSub.Unsubscribe()
	d.unregisterStatus()
	d.wg.Wait()

	return nil
}

// Stop 停止设备
func (d *Device) Stop() {
	d.cancel()
	d.nc.Drain()
	d.nc.Close()
}

// registerStatus 注册设备状态
// 对应项目中的 RegisterDevice
func (d *Device) registerStatus() error {
	status := protocol.NewDeviceStatus(d.id, d.accountID, d.nodeType)
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("序列化状态失败: %w", err)
	}

	_, err = d.kv.Put(d.ctx, d.id, data)
	if err != nil {
		return fmt.Errorf("注册状态失败: %w", err)
	}

	log.Printf("[device-%s] 状态已注册", d.id)
	return nil
}

// unregisterStatus 注销设备状态
// 对应项目中的 UnregisterDevice
func (d *Device) unregisterStatus() {
	// 使用 Delete 或更新状态为 offline
	status := &protocol.DeviceStatus{
		DeviceID:      d.id,
		Status:        "offline",
		LastHeartbeat: time.Now().Unix(),
	}
	data, _ := json.Marshal(status)
	d.kv.Put(d.ctx, d.id, data)

	log.Printf("[device-%s] 状态已注销", d.id)
}

// subscribeCommand 订阅命令
// 对应项目中的 WebSocket 消息读取循环
func (d *Device) subscribeCommand() (*nats.Subscription, error) {
	subject := protocol.GetCommandSubject(d.id)

	sub, err := d.nc.Subscribe(subject, func(msg *nats.Msg) {
		var req protocol.CommandRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("[device-%s] 解析命令失败: %v", d.id, err)
			d.sendErrorResponse(msg.Reply, req.MsgID, "无效的命令格式")
			return
		}

		log.Printf("[device-%s] 收到命令: msg_id=%s, query=%s", d.id, req.MsgID, req.Query)

		// 处理命令（流式响应）
		d.handleCommand(msg.Reply, &req)
	})

	if err != nil {
		return nil, fmt.Errorf("订阅命令失败: %w", err)
	}

	return sub, nil
}

// subscribeCancel 订阅取消
func (d *Device) subscribeCancel() (*nats.Subscription, error) {
	subject := protocol.GetCancelSubject(d.id)

	sub, err := d.nc.Subscribe(subject, func(msg *nats.Msg) {
		var req protocol.CancelRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("[device-%s] 解析取消请求失败: %v", d.id, err)
			return
		}

		log.Printf("[device-%s] 收到取消请求: msg_id=%s", d.id, req.MsgID)

		// TODO: 实现取消逻辑
		// 可以维护一个 map[string]context.CancelFunc 来取消正在执行的命令
	})

	if err != nil {
		return nil, fmt.Errorf("订阅取消失败: %w", err)
	}

	return sub, nil
}

// handleCommand 处理命令
// 对应项目中的 query 处理逻辑
func (d *Device) handleCommand(replyTo string, req *protocol.CommandRequest) {
	// 模拟流式响应
	// 实际项目中，这里会调用设备的具体处理逻辑

	chunks := []string{
		"正在处理您的请求...\n",
		"分析中...\n",
		"生成响应...\n",
		"完成。",
	}

	// 发送流式响应
	for i, chunk := range chunks {
		select {
		case <-d.ctx.Done():
			return
		default:
		}

		resp := protocol.NewChunkResponse(req.MsgID, i+1, []byte(chunk))
		data, _ := json.Marshal(resp)

		if err := d.nc.Publish(replyTo, data); err != nil {
			log.Printf("[device-%s] 发送响应失败: %v", d.id, err)
			return
		}

		// 模拟处理延迟
		time.Sleep(100 * time.Millisecond)
	}

	// 发送完成响应
	doneResp := protocol.NewDoneResponse(req.MsgID)
	data, _ := json.Marshal(doneResp)
	d.nc.Publish(replyTo, data)

	log.Printf("[device-%s] 命令完成: msg_id=%s", d.id, req.MsgID)
}

// sendErrorResponse 发送错误响应
func (d *Device) sendErrorResponse(replyTo, msgID, errMsg string) {
	resp := protocol.NewErrorResponse(msgID, errMsg)
	data, _ := json.Marshal(resp)
	d.nc.Publish(replyTo, data)
}

// heartbeatLoop 心跳循环
// 对应项目中的 StartHeartbeat
func (d *Device) heartbeatLoop() {
	defer d.wg.Done()

	ticker := time.NewTicker(d.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.updateHeartbeat()
		}
	}
}

// updateHeartbeat 更新心跳
func (d *Device) updateHeartbeat() {
	// 获取当前状态
	entry, err := d.kv.Get(d.ctx, d.id)
	if err != nil {
		// 如果不存在，重新注册
		d.registerStatus()
		return
	}

	var status protocol.DeviceStatus
	if err := json.Unmarshal(entry.Value(), &status); err != nil {
		d.registerStatus()
		return
	}

	// 更新心跳时间
	status.UpdateHeartbeat()
	data, _ := json.Marshal(status)

	_, err = d.kv.Put(d.ctx, d.id, data)
	if err != nil {
		log.Printf("[device-%s] 更新心跳失败: %v", d.id, err)
	}
}

// PublishTelemetry 发布遥测数据
// 对应项目中的 telemetry 消息
func (d *Device) PublishTelemetry(typ string, data []byte) error {
	telemetry := protocol.NewTelemetry(d.id, typ, data)
	msg, _ := json.Marshal(telemetry)

	subject := protocol.GetTelemetrySubject(d.id)
	return d.nc.Publish(subject, msg)
}

// PublishStatus 发布状态变更
func (d *Device) PublishStatus(status, reason string) error {
	msg := protocol.NewStatusMessage(d.id, status, reason)
	data, _ := json.Marshal(msg)

	subject := protocol.GetStatusSubject(d.id)
	return d.nc.Publish(subject, data)
}