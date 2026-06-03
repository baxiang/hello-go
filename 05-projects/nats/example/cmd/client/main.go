// Package main 测试客户端
// 用于测试 Hub 的执行命令接口
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	// 解析命令行参数
	hubURL := flag.String("hub", "http://localhost:8080", "Hub 服务地址")
	deviceID := flag.String("device", "device-001", "设备 ID")
	query := flag.String("query", "你好，请介绍一下自己", "查询内容")
	flag.Parse()

	log.Printf("发送命令: device=%s, query=%s", *deviceID, *query)

	// 构建请求
	reqBody := map[string]string{"query": *query}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/api/v1/device/execute?device_id=%s", *hubURL, *deviceID)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("请求失败: %s - %s", resp.Status, string(respBody))
	}

	log.Println("开始接收流式响应:")

	// 读取 SSE 流
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("读取错误: %v", err)
			break
		}

		line = trimNewline(line)
		if line == "" {
			continue
		}

		// 解析 SSE 数据
		if len(line) > 6 && line[:6] == "data: " {
			data := line[6:]
			if data == "[DONE]" {
				log.Println("\n--- 响应完成 ---")
				break
			}

			var resp struct {
				Type      string          `json:"type"`
				Seq       int             `json:"seq"`
				Data      json.RawMessage `json:"data"`
				Error     string          `json:"error"`
				Timestamp int64           `json:"timestamp"`
			}
			if err := json.Unmarshal([]byte(data), &resp); err != nil {
				log.Printf("解析响应失败: %v", err)
				continue
			}

			if resp.Error != "" {
				log.Printf("错误: %s", resp.Error)
				break
			}

			fmt.Print(string(resp.Data))
		}
	}
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] == '\r' {
		s = s[:len(s)-1]
	}
	return s
}