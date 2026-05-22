// Package main 设备端主程序
// 模拟设备连接到 NATS 并接收命令
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"nats-mvp-example/internal/config"
	"nats-mvp-example/internal/device"
)

func main() {
	// 解析命令行参数
	deviceID := flag.String("id", "device-001", "设备 ID")
	accountID := flag.String("account", "account-001", "账号 ID")
	nodeType := flag.String("type", "agent", "节点类型")
	natsURL := flag.String("nats", config.DefaultNATSURL, "NATS 服务器地址")
	flag.Parse()

	log.Printf("启动设备: id=%s, account=%s, type=%s", *deviceID, *accountID, *nodeType)

	// 创建设备配置
	cfg := &config.DeviceConfig{
		ID:               *deviceID,
		AccountID:        *accountID,
		NodeType:         *nodeType,
		NATSURL:          *natsURL,
		HeartbeatInterval: 30e9, // 30s
	}

	// 创建设备实例
	d, err := device.New(cfg)
	if err != nil {
		log.Fatalf("创建设备失败: %v", err)
	}

	// 启动设备
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := d.Start(); err != nil {
			log.Printf("设备运行错误: %v", err)
			cancel()
		}
	}()

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("收到退出信号")
	case <-ctx.Done():
	}

	// 停止设备
	d.Stop()
	log.Println("设备已停止")
}