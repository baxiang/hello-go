// Package main 端到端测试
// 直接使用 NATS 进行通信，不经过 HTTP
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nats-mvp-example/internal/config"
	"nats-mvp-example/internal/device"
	"nats-mvp-example/internal/hub"
	"nats-mvp-example/internal/protocol"
)

func main() {
	// 解析命令行参数
	mode := flag.String("mode", "test", "运行模式: device, hub, test")
	deviceID := flag.String("device", "device-001", "设备 ID")
	natsURL := flag.String("nats", config.DefaultNATSURL, "NATS 服务器地址")
	flag.Parse()

	switch *mode {
	case "device":
		runDevice(*deviceID, *natsURL)
	case "hub":
		runHub(*natsURL)
	case "test":
		runTest(*natsURL)
	default:
		log.Fatalf("未知模式: %s", *mode)
	}
}

// runDevice 运行设备端
func runDevice(deviceID, natsURL string) {
	log.Printf("启动设备: id=%s", deviceID)

	cfg := &config.DeviceConfig{
		ID:               deviceID,
		AccountID:        "account-001",
		NodeType:         "agent",
		NATSURL:          natsURL,
		HeartbeatInterval: 30 * time.Second,
	}

	d, err := device.New(cfg)
	if err != nil {
		log.Fatalf("创建设备失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := d.Start(); err != nil {
			log.Printf("设备运行错误: %v", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("收到退出信号")
	case <-ctx.Done():
	}

	d.Stop()
	log.Println("设备已停止")
}

// runHub 运行 Hub 服务
func runHub(natsURL string) {
	log.Printf("启动 Hub 服务")

	cfg := &config.HubConfig{
		NATSURL:          natsURL,
		ExecuteTimeout:   60 * time.Second,
		HeartbeatTimeout: 90 * time.Second,
	}

	h, err := hub.New(cfg)
	if err != nil {
		log.Fatalf("创建 Hub 失败: %v", err)
	}

	if err := h.Start(); err != nil {
		log.Fatalf("启动 Hub 失败: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	h.Stop()
	log.Println("Hub 已停止")
}

// runTest 运行端到端测试
func runTest(natsURL string) {
	log.Println("=== 端到端测试 ===")

	ctx := context.Background()

	// 1. 创建 Hub
	hubCfg := &config.HubConfig{
		NATSURL:          natsURL,
		ExecuteTimeout:   30 * time.Second,
		HeartbeatTimeout: 90 * time.Second,
	}

	h, err := hub.New(hubCfg)
	if err != nil {
		log.Fatalf("创建 Hub 失败: %v", err)
	}
	defer h.Stop()

	if err := h.Start(); err != nil {
		log.Fatalf("启动 Hub 失败: %v", err)
	}

	// 2. 创建设备
	deviceCfg := &config.DeviceConfig{
		ID:               "test-device-001",
		AccountID:        "account-001",
		NodeType:         "agent",
		NATSURL:          natsURL,
		HeartbeatInterval: 30 * time.Second,
	}

	d, err := device.New(deviceCfg)
	if err != nil {
		log.Fatalf("创建设备失败: %v", err)
	}
	defer d.Stop()

	// 启动设备（后台）
	_, deviceCancel := context.WithCancel(context.Background())
	go func() {
		if err := d.Start(); err != nil {
			log.Printf("设备运行错误: %v", err)
		}
	}()
	defer deviceCancel()

	// 等待设备注册
	time.Sleep(2 * time.Second)

	// 3. 检查设备在线
	log.Println("\n--- 测试 1: 检查设备在线 ---")
	online := h.IsDeviceOnline(ctx, "test-device-001")
	log.Printf("设备在线: %v", online)

	if !online {
		log.Println("警告: 设备未在线，可能是 NATS 连接问题")
	}

	// 4. 获取设备状态
	log.Println("\n--- 测试 2: 获取设备状态 ---")
	status, err := h.GetDeviceStatus(ctx, "test-device-001")
	if err != nil {
		log.Printf("获取状态失败: %v", err)
	} else {
		log.Printf("设备状态: id=%s, status=%s, node_type=%s", status.DeviceID, status.Status, status.NodeType)
	}

	// 5. 列出在线设备
	log.Println("\n--- 测试 3: 列出在线设备 ---")
	devices, err := h.ListOnlineDevices(ctx)
	if err != nil {
		log.Printf("列出设备失败: %v", err)
	} else {
		log.Printf("在线设备数: %d", len(devices))
		for _, d := range devices {
			log.Printf("  - %s (%s)", d.DeviceID, d.Status)
		}
	}

	// 6. 执行命令
	log.Println("\n--- 测试 4: 执行命令 ---")
	if online {
		respCh, err := h.ExecuteCommand(ctx, "test-device-001", "你好，请介绍一下自己")
		if err != nil {
			log.Printf("执行命令失败: %v", err)
		} else {
			log.Println("命令已发送，等待响应...")
			for resp := range respCh {
				log.Printf("响应: type=%s, seq=%d, data=%s", resp.Type, resp.Seq, string(resp.Data))
				if resp.Type == protocol.TypeResponseDone || resp.Type == protocol.TypeError {
					break
				}
			}
		}
	} else {
		log.Println("跳过命令执行测试（设备不在线）")
	}

	// 7. 监听设备状态变更
	log.Println("\n--- 测试 5: 监听设备状态 ---")
	statusCh, err := h.WatchDeviceStatus(ctx)
	if err != nil {
		log.Printf("创建状态监听失败: %v", err)
	} else {
		log.Println("开始监听设备状态变更（5秒）...")
		go func() {
			for status := range statusCh {
				log.Printf("状态变更: device=%s, status=%s", status.DeviceID, status.Status)
			}
		}()

		time.Sleep(5 * time.Second)
	}

	log.Println("\n=== 测试完成 ===")
}

// 辅助函数
func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}