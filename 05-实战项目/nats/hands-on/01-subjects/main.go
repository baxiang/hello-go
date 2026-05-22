// 01 - Subject 寻址实践
//
// 理论知识：参见 ../02-subjects.md
//
// 核心概念：
// - Subject 是消息的地址，用 . 分隔层级
// - * 匹配单个层级
// - > 匹配剩余所有层级（只能在末尾）
// - 大小写敏感
//
// 运行方式：
//   go run ./01-subjects/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NATS 服务器地址
const natsURL = "nats://localhost:4222"

func main() {
	// 连接 NATS
	nc, err := nats.Connect(natsURL,
		nats.Name("hands-on-01-subjects"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Drain()

	fmt.Println("✅ 已连接到 NATS:", nc.ConnectedUrl())
	fmt.Println()

	// ========================================
	// 第一部分：精确匹配 vs 通配符匹配
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第一部分：精确匹配 vs 通配符匹配")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// 订阅者 1：精确匹配 orders.created
	sub1, _ := nc.Subscribe("orders.created", func(msg *nats.Msg) {
		fmt.Printf("  📨 [精确匹配 orders.created] 收到: %s\n", string(msg.Data))
	})

	// 订阅者 2：单层通配符 orders.*
	sub2, _ := nc.Subscribe("orders.*", func(msg *nats.Msg) {
		fmt.Printf("  📨 [单层通配 orders.*] 收到: subject=%s data=%s\n", msg.Subject, string(msg.Data))
	})

	// 订阅者 3：多层通配符 orders.>
	sub3, _ := nc.Subscribe("orders.>", func(msg *nats.Msg) {
		fmt.Printf("  📨 [多层通配 orders.>] 收到: subject=%s data=%s\n", msg.Subject, string(msg.Data))
	})

	nc.Flush() // 确保订阅已生效

	// 测试消息
	testCases := []struct {
		subject string
		data    string
	}{
		{"orders.created", `{"id":"001"}`},        // 匹配全部 3 个订阅
		{"orders.cancelled", `{"id":"002"}`},      // 匹配 orders.* 和 orders.>
		{"orders.payment.completed", `{"id":"003"}`}, // 只匹配 orders.>
		{"orders.us.east.created", `{"id":"004"}`},   // 只匹配 orders.>
	}

	for _, tc := range testCases {
		fmt.Printf("📤 发布: subject=%s\n", tc.subject)
		nc.Publish(tc.subject, []byte(tc.data))
		time.Sleep(100 * time.Millisecond)
		fmt.Println()
	}

	sub1.Unsubscribe()
	sub2.Unsubscribe()
	sub3.Unsubscribe()

	// ========================================
	// 第二部分：通配符匹配规则详解
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第二部分：通配符匹配规则详解")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("规则说明：")
	fmt.Println("  * 匹配单个层级，不能跨越 .")
	fmt.Println("  > 匹配剩余所有层级，只能在末尾")
	fmt.Println()

	// 匹配表格演示
	fmt.Println("匹配测试表：")
	fmt.Println("┌─────────────────────────┬───────────┬───────────┬─────────────────┐")
	fmt.Println("│ 发布的 Subject          │ orders.*  │ orders.>  │ orders.*.created│")
	fmt.Println("├─────────────────────────┼───────────┼───────────┼─────────────────┤")

	tests := []struct {
		subject string
		star    bool
		gt      bool
		starEnd bool
	}{
		{"orders.created", true, true, false},
		{"orders.us", true, true, false},
		{"orders.us.created", false, true, true},
		{"orders.us.east.created", false, true, false},
	}

	for _, t := range tests {
		fmt.Printf("│ %-23s │ %-9s │ %-9s │ %-15s │\n",
			t.subject,
			boolStr(t.star),
			boolStr(t.gt),
			boolStr(t.starEnd),
		)
	}
	fmt.Println("└─────────────────────────┴───────────┴───────────┴─────────────────┘")
	fmt.Println()

	// ========================================
	// 第三部分：实际业务场景 Subject 设计
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第三部分：实际业务场景 Subject 设计")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("推荐命名规范：<domain>.<entity>.<action>")
	fmt.Println()

	// 电商订单系统示例
	fmt.Println("📦 电商订单系统：")
	orderSubjects := []string{
		"orders.created",
		"orders.payment.completed",
		"orders.payment.failed",
		"orders.shipment.dispatched",
		"orders.shipment.delivered",
		"orders.cancelled",
	}
	for _, s := range orderSubjects {
		fmt.Printf("  %s\n", s)
	}
	fmt.Println()

	// IoT 设备管理示例
	fmt.Println("🔧 IoT 设备管理：")
	iotSubjects := []string{
		"iot.device.{device-id}.telemetry",
		"iot.device.{device-id}.command",
		"iot.device.{device-id}.status",
		"iot.alert.critical",
		"iot.alert.warning",
	}
	for _, s := range iotSubjects {
		fmt.Printf("  %s\n", s)
	}
	fmt.Println()

	// 微服务 RPC 示例
	fmt.Println("🔌 微服务 RPC：")
	rpcSubjects := []string{
		"rpc.user.get",
		"rpc.order.create",
		"rpc.payment.process",
	}
	for _, s := range rpcSubjects {
		fmt.Printf("  %s\n", s)
	}
	fmt.Println()

	// ========================================
	// 第四部分：反模式警告
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第四部分：反模式警告")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("❌ 错误示例：")
	fmt.Println()
	fmt.Println("  1. 使用下划线代替层级分隔：")
	fmt.Println("     orders_payment_completed  ← 错误，无法用通配符匹配")
	fmt.Println("     orders.payment.completed  ← 正确")
	fmt.Println()
	fmt.Println("  2. 层级过深（> 8 层）：")
	fmt.Println("     com.company.product.region.datacenter.service.module.action")
	fmt.Println("     product.region.service.action  ← 精简")
	fmt.Println()
	fmt.Println("  3. ID 作为顶层：")
	fmt.Println("     {device-id}.telemetry  ← 百万设备 = 百万顶层")
	fmt.Println("     iot.device.{device-id}.telemetry  ← 正确")
	fmt.Println()
	fmt.Println("  4. 大小写混用：")
	fmt.Println("     Orders.Created ≠ orders.created ≠ ORDERS.CREATED")
	fmt.Println("     orders.created  ← 统一小写")
	fmt.Println()

	// ========================================
	// 第五部分：实战练习
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第五部分：实战练习")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("练习：设计一个设备管理系统的 Subject 命名")
	fmt.Println()
	fmt.Println("需求：")
	fmt.Println("  - 设备上线/下线通知")
	fmt.Println("  - 设备遥测数据上报（温度、湿度、位置）")
	fmt.Println("  - 设备命令下发（重启、配置更新）")
	fmt.Println("  - 设备告警（严重、警告、信息）")
	fmt.Println()
	fmt.Print("按 Enter 键查看参考答案...")
	// fmt.Scanln() // 自动运行模式，跳过等待

	fmt.Println()
	fmt.Println("参考答案：")
	answer := []string{
		"device.{id}.connected",
		"device.{id}.disconnected",
		"telemetry.{id}.temperature",
		"telemetry.{id}.humidity",
		"telemetry.{id}.location",
		"command.{id}.reboot",
		"command.{id}.config",
		"alert.{id}.critical",
		"alert.{id}.warning",
		"alert.{id}.info",
	}
	for _, s := range answer {
		fmt.Printf("  ✅ %s\n", s)
	}
	fmt.Println()

	// 订阅所有设备告警的示例
	fmt.Println("订阅所有设备告警：")
	fmt.Println("  nc.Subscribe(\"alert.*.>\", handler)")
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第一章学习完成！")
	fmt.Println("下一篇：02-pub-sub - Pub/Sub 发布订阅")
	fmt.Println("════════════════════════════════════════════════════════════")
}

func boolStr(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}

func init() {
	// 禁止 go vet 报告未使用的变量
	_ = strings.NewReader("")
	_ = os.Args
}