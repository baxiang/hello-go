// Package client 实现客户端
// 对应项目中的 HTTP SSE 客户端
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"nats-mvp-example/internal/config"
)

// Client 客户端
type Client struct {
	hubURL string
	client *http.Client
}

// New 创建客户端
func New(cfg *config.ClientConfig) *Client {
	return &Client{
		hubURL: cfg.HubURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ExecuteRequest 执行请求
type ExecuteRequest struct {
	Query string `json:"query"`
}

// ExecuteResponse 执行响应
type ExecuteResponse struct {
	Content string `json:"content,omitempty"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

// Execute 执行命令（SSE 流式）
// 对应项目中的 POST /api/v1/device/execute
func (c *Client) Execute(ctx context.Context, deviceID, query string) (<-chan *ExecuteResponse, error) {
	// 构建请求
	reqBody := ExecuteRequest{Query: query}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/api/v1/device/execute?device_id=%s", c.hubURL, deviceID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("请求失败: %s", resp.Status)
	}

	// 创建响应通道
	respCh := make(chan *ExecuteResponse, 10)

	go func() {
		defer close(respCh)
		defer resp.Body.Close()

		// 读取 SSE 流
		decoder := NewSSEDecoder(resp.Body)
		for {
			event, err := decoder.Decode()
			if err != nil {
				if err != io.EOF {
					log.Printf("[client] 读取 SSE 失败: %v", err)
				}
				return
			}

			if event.Data == "[DONE]" {
				respCh <- &ExecuteResponse{Done: true}
				return
			}

			var execResp ExecuteResponse
			if err := json.Unmarshal([]byte(event.Data), &execResp); err != nil {
				log.Printf("[client] 解析响应失败: %v", err)
				continue
			}

			respCh <- &execResp
		}
	}()

	return respCh, nil
}

// SSEEvent SSE 事件
type SSEEvent struct {
	Event string
	Data  string
}

// SSEDecoder SSE 解码器
type SSEDecoder struct {
	reader io.Reader
}

// NewSSEDecoder 创建 SSE 解码器
func NewSSEDecoder(reader io.Reader) *SSEDecoder {
	return &SSEDecoder{reader: reader}
}

// Decode 解码 SSE 事件
func (d *SSEDecoder) Decode() (*SSEEvent, error) {
	// 简化的 SSE 解析
	// 实际实现需要处理多行事件
	buf := make([]byte, 4096)
	n, err := io.ReadFull(d.reader, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	data := string(buf[:n])
	lines := strings.Split(data, "\n")

	var event SSEEvent
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			event.Data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		}
	}

	return &event, nil
}

// ExecuteSync 同步执行命令
func (c *Client) ExecuteSync(ctx context.Context, deviceID, query string) (string, error) {
	respCh, err := c.Execute(ctx, deviceID, query)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for resp := range respCh {
		if resp.Error != "" {
			return "", fmt.Errorf(resp.Error)
		}
		if resp.Done {
			break
		}
		result.WriteString(resp.Content)
	}

	return result.String(), nil
}

// NATSClient 直接使用 NATS 的客户端
// 用于测试 NATS 直连场景
type NATSClient struct {
	hub *NATSHubClient
}

// NATSHubClient NATS Hub 客户端
type NATSHubClient struct {
	// 这里应该引用 hub.Hub，但为了避免循环依赖，使用接口
}

// NewNATSClient 创建 NATS 客户端
func NewNATSClient(cfg *config.ClientConfig) (*NATSClient, error) {
	return &NATSClient{}, nil
}

// Execute 执行命令
func (c *NATSClient) Execute(ctx context.Context, deviceID, query string) (<-chan *ExecuteResponse, error) {
	// 直接通过 NATS 发送命令
	// 这里需要引用 hub.Hub 的 ExecuteCommand 方法
	return nil, fmt.Errorf("not implemented")
}