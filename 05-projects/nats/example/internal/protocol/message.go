// Package protocol 定义 NATS 消息协议
// 与项目 WebSocket 协议保持语义一致
package protocol

import (
	"encoding/json"
	"time"
)

// 消息类型常量
const (
	// 命令类型
	TypeCommand = "command" // 执行命令
	TypeCancel  = "cancel"  // 取消请求

	// 响应类型
	TypeResponseChunk = "response_chunk" // 流式响应块
	TypeResponseDone  = "response_done"  // 响应完成
	TypeError         = "error"          // 错误响应

	// 状态类型
	TypeStatusOnline  = "status_online"  // 设备上线
	TypeStatusOffline = "status_offline" // 设备下线
	TypeHeartbeat     = "heartbeat"      // 心跳

	// 遥测类型
	TypeTelemetry = "telemetry" // 遥测数据
)

// Subject 命名规范
const (
	// 设备命令 Subject
	// 格式: device.{device_id}.command
	// 用途: Hub 向设备发送执行命令
	SubjectCommandFormat = "device.%s.command"

	// 设备取消 Subject
	// 格式: device.{device_id}.cancel
	// 用途: Hub 向设备发送取消请求
	SubjectCancelFormat = "device.%s.cancel"

	// 设备响应 Subject
	// 格式: device.{device_id}.response
	// 用途: 设备向 Hub 发送命令响应
	SubjectResponseFormat = "device.%s.response"

	// 设备遥测 Subject
	// 格式: device.{device_id}.telemetry
	// 用途: 设备向 Hub 发送遥测数据
	SubjectTelemetryFormat = "device.%s.telemetry"

	// 设备状态 Subject
	// 格式: device.{device_id}.status
	// 用途: 设备向 Hub 发送状态变更
	SubjectStatusFormat = "device.%s.status"

	// Hub 订阅所有设备响应
	// 格式: device.*.response
	SubjectAllResponse = "device.*.response"

	// Hub 订阅所有设备遥测
	// 格式: device.*.telemetry
	SubjectAllTelemetry = "device.*.telemetry"

	// Hub 订阅所有设备状态
	// 格式: device.*.status
	SubjectAllStatus = "device.*.status"
)

// GetCommandSubject 获取设备命令 Subject
func GetCommandSubject(deviceID string) string {
	return formatSubject(SubjectCommandFormat, deviceID)
}

// GetCancelSubject 获取设备取消 Subject
func GetCancelSubject(deviceID string) string {
	return formatSubject(SubjectCancelFormat, deviceID)
}

// GetResponseSubject 获取设备响应 Subject
func GetResponseSubject(deviceID string) string {
	return formatSubject(SubjectResponseFormat, deviceID)
}

// GetTelemetrySubject 获取设备遥测 Subject
func GetTelemetrySubject(deviceID string) string {
	return formatSubject(SubjectTelemetryFormat, deviceID)
}

// GetStatusSubject 获取设备状态 Subject
func GetStatusSubject(deviceID string) string {
	return formatSubject(SubjectStatusFormat, deviceID)
}

func formatSubject(format, deviceID string) string {
	return format[:len(format)-4] + deviceID + format[len(format)-4:]
}

// CommandRequest 命令请求
// 对应 WebSocket 协议: {"type":"query","msg_id":"xxx","payload":{"query":"..."}}
type CommandRequest struct {
	MsgID     string          `json:"msg_id"`               // 消息唯一标识
	Query     string          `json:"query"`                // 查询内容
	Timeout   int             `json:"timeout,omitempty"`    // 超时时间（秒）
	Timestamp int64           `json:"timestamp"`            // 请求时间戳
	Metadata  json.RawMessage `json:"metadata,omitempty"`   // 元数据
}

// NewCommandRequest 创建命令请求
func NewCommandRequest(query string) *CommandRequest {
	return &CommandRequest{
		MsgID:     generateMsgID(),
		Query:     query,
		Timestamp: time.Now().Unix(),
	}
}

// CancelRequest 取消请求
type CancelRequest struct {
	MsgID     string `json:"msg_id"`              // 要取消的消息 ID
	Reason    string `json:"reason,omitempty"`   // 取消原因
	Timestamp int64  `json:"timestamp"`          // 请求时间戳
}

// NewCancelRequest 创建取消请求
func NewCancelRequest(msgID string) *CancelRequest {
	return &CancelRequest{
		MsgID:     msgID,
		Timestamp: time.Now().Unix(),
	}
}

// Response 命令响应
// 对应 WebSocket 协议: {"type":"response_chunk","msg_id":"xxx","payload":"..."}
type Response struct {
	MsgID     string          `json:"msg_id"`              // 对应的请求 ID
	Type      string          `json:"type"`                // 响应类型
	Seq       int             `json:"seq,omitempty"`       // 序列号（流式响应）
	Data      json.RawMessage `json:"data"`                // 响应数据
	Error     string          `json:"error,omitempty"`     // 错误信息
	Timestamp int64           `json:"timestamp"`           // 响应时间戳
}

// NewChunkResponse 创建流式响应块
func NewChunkResponse(msgID string, seq int, data []byte) *Response {
	// 将数据转换为 JSON 格式
	jsonData, _ := json.Marshal(string(data))
	return &Response{
		MsgID:     msgID,
		Type:      TypeResponseChunk,
		Seq:       seq,
		Data:      json.RawMessage(jsonData),
		Timestamp: time.Now().Unix(),
	}
}

// NewDoneResponse 创建完成响应
func NewDoneResponse(msgID string) *Response {
	return &Response{
		MsgID:     msgID,
		Type:      TypeResponseDone,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(msgID, errMsg string) *Response {
	return &Response{
		MsgID:     msgID,
		Type:      TypeError,
		Error:     errMsg,
		Timestamp: time.Now().Unix(),
	}
}

// DeviceStatus 设备状态
// 用于 KV Store 存储
type DeviceStatus struct {
	DeviceID      string `json:"device_id"`                // 设备 ID
	AccountID     string `json:"account_id"`               // 账号 ID
	NodeType      string `json:"node_type"`                // 节点类型
	Status        string `json:"status"`                   // online/offline
	LastHeartbeat int64  `json:"last_heartbeat"`            // 最后心跳时间
	Version       string `json:"version,omitempty"`        // 设备版本
	IPAddress     string `json:"ip_address,omitempty"`     // IP 地址
}

// NewDeviceStatus 创建设备状态
func NewDeviceStatus(deviceID, accountID, nodeType string) *DeviceStatus {
	return &DeviceStatus{
		DeviceID:      deviceID,
		AccountID:     accountID,
		NodeType:      nodeType,
		Status:        "online",
		LastHeartbeat: time.Now().Unix(),
	}
}

// UpdateHeartbeat 更新心跳时间
func (s *DeviceStatus) UpdateHeartbeat() {
	s.LastHeartbeat = time.Now().Unix()
}

// IsOnline 检查是否在线
func (s *DeviceStatus) IsOnline(timeout int64) bool {
	if s.Status != "online" {
		return false
	}
	return time.Now().Unix()-s.LastHeartbeat < timeout
}

// Telemetry 遥测数据
type Telemetry struct {
	DeviceID  string          `json:"device_id"`           // 设备 ID
	Type      string          `json:"type"`                // 遥测类型
	Data      json.RawMessage `json:"data"`                // 遥测数据
	Timestamp int64           `json:"timestamp"`           // 时间戳
}

// NewTelemetry 创建遥测数据
func NewTelemetry(deviceID, typ string, data []byte) *Telemetry {
	return &Telemetry{
		DeviceID:  deviceID,
		Type:      typ,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// StatusMessage 状态变更消息
type StatusMessage struct {
	DeviceID string `json:"device_id"`         // 设备 ID
	Status   string `json:"status"`            // online/offline
	Reason   string `json:"reason,omitempty"`  // 变更原因
	Timestamp int64 `json:"timestamp"`         // 时间戳
}

// NewStatusMessage 创建状态消息
func NewStatusMessage(deviceID, status, reason string) *StatusMessage {
	return &StatusMessage{
		DeviceID:  deviceID,
		Status:    status,
		Reason:    reason,
		Timestamp: time.Now().Unix(),
	}
}