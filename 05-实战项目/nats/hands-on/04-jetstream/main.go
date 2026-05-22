// 04 - JetStream 实践
//
// 理论知识：参见 ../05-jetstream-streams.md 和 ../06-jetstream-consumers.md
//
// 核心概念：
// - JetStream 是 NATS 的持久化引擎
// - Stream：持久化消息日志
// - Consumer：读取 Stream 的游标
// - At Least Once：消息至少投递一次
// - ACK 机制：确认消息处理成功
//
// 运行方式：
//   go run ./04-jetstream/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const natsURL = "nats://localhost:4222"

// 订单事件
type OrderEvent struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}

func main() {
	nc, err := nats.Connect(natsURL,
		nats.Name("hands-on-04-jetstream"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Drain()

	fmt.Println("✅ 已连接到 NATS:", nc.ConnectedUrl())
	fmt.Println()

	// 创建 JetStream Context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("创建 JetStream 失败: %v", err)
	}

	ctx := context.Background()

	// ========================================
	// 第一部分：创建 Stream
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第一部分：创建 Stream")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("Stream：持久化消息日志")
	fmt.Println()

	// 创建或更新 Stream
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        "ORDERS_HANDSON",
		Description: "订单事件流（学习示例）",
		Subjects:    []string{"orders.handson.>"},
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      24 * time.Hour, // 保留 24 小时
		MaxBytes:    10 * 1024 * 1024, // 最大 10MB
		Replicas:    1, // 开发环境用 1，生产用 3
	})
	if err != nil {
		log.Fatalf("创建 Stream 失败: %v", err)
	}

	info, _ := stream.Info(ctx)
	fmt.Printf("✅ Stream 已创建: %s\n", info.Config.Name)
	fmt.Printf("   Subjects: %v\n", info.Config.Subjects)
	fmt.Printf("   Max Age: %v\n", info.Config.MaxAge)
	fmt.Println()

	// ========================================
	// 第二部分：发布消息
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第二部分：发布消息到 Stream")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// 同步发布（等待 ACK）
	fmt.Println("📝 同步发布（等待持久化确认）：")
	for i := 1; i <= 3; i++ {
		order := OrderEvent{
			OrderID: fmt.Sprintf("ORD-%03d", i),
			Amount:  float64(i * 100),
			Status:  "created",
		}
		data, _ := json.Marshal(order)

		ack, err := js.Publish(ctx, "orders.handson.created", data)
		if err != nil {
			fmt.Printf("  ❌ 发布失败: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ 已持久化: seq=%d stream=%s\n", ack.Sequence, ack.Stream)
	}
	fmt.Println()

	// 查看 Stream 状态
	info, _ = stream.Info(ctx)
	fmt.Printf("📊 Stream 状态: 消息数=%d 字节数=%d\n", info.State.Msgs, info.State.Bytes)
	fmt.Println()

	// ========================================
	// 第三部分：Pull Consumer 消费
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第三部分：Pull Consumer 消费")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("Pull Consumer：客户端主动拉取消息")
	fmt.Println()

	// 创建 Consumer
	consumer, err := js.CreateOrUpdateConsumer(ctx, "ORDERS_HANDSON", jetstream.ConsumerConfig{
		Name:          "order-processor",
		Durable:       "order-processor",
		Description:   "订单处理器",
		FilterSubject: "orders.handson.>",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    3,
	})
	if err != nil {
		log.Fatalf("创建 Consumer 失败: %v", err)
	}

	fmt.Printf("✅ Consumer 已创建: %s\n", consumer.CachedInfo().Name)
	fmt.Println()

	// 拉取消息
	fmt.Println("📝 拉取消息：")
	msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(2*time.Second))
	if err != nil {
		fmt.Printf("  ⚠️ Fetch 错误: %v\n", err)
	} else {
		count := 0
		for msg := range msgs.Messages() {
			count++
			var order OrderEvent
			json.Unmarshal(msg.Data(), &order)

			meta, _ := msg.Metadata()
			fmt.Printf("  📨 消息 %d: seq=%d order_id=%s amount=%.0f\n",
				count, meta.Sequence.Stream, order.OrderID, order.Amount)

			// 确认消息
			msg.Ack()
		}
		if count == 0 {
			fmt.Println("  （没有待消费的消息）")
		}
	}
	fmt.Println()

	// ========================================
	// 第四部分：ACK 策略详解
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第四部分：ACK 策略详解")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("ACK 策略：")
	fmt.Println("┌─────────────────┬─────────────────────────────────────────┐")
	fmt.Println("│ 策略            │ 说明                                    │")
	fmt.Println("├─────────────────┼─────────────────────────────────────────┤")
	fmt.Println("│ AckExplicit     │ 需要显式 ACK，每条消息单独确认（推荐）   │")
	fmt.Println("│ AckAll          │ ACK 一条消息，之前的全部确认             │")
	fmt.Println("│ AckNone         │ 不需要 ACK                              │")
	fmt.Println("└─────────────────┴─────────────────────────────────────────┘")
	fmt.Println()

	fmt.Println("ACK 操作：")
	fmt.Println("┌──────────────────┬─────────────────────────────────────────┐")
	fmt.Println("│ 方法             │ 说明                                    │")
	fmt.Println("├──────────────────┼─────────────────────────────────────────┤")
	fmt.Println("│ msg.Ack()        │ 确认处理成功                            │")
	fmt.Println("│ msg.Nak()        │ 处理失败，立即重投                      │")
	fmt.Println("│ msg.NakWithDelay │ 处理失败，延迟重投                      │")
	fmt.Println("│ msg.InProgress() │ 正在处理，延长 AckWait                  │")
	fmt.Println("│ msg.Term()       │ 永久丢弃，不再重投                      │")
	fmt.Println("└──────────────────┴─────────────────────────────────────────┘")
	fmt.Println()

	// ========================================
	// 第五部分：消息去重
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第五部分：消息去重")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// 更新 Stream 配置启用去重
	stream, _ = js.UpdateStream(ctx, jetstream.StreamConfig{
		Name:       "ORDERS_HANDSON",
		Subjects:   []string{"orders.handson.>"},
		MaxAge:     24 * time.Hour,
		MaxBytes:   10 * 1024 * 1024,
		Replicas:   1,
		Duplicates: 5 * time.Minute, // 5 分钟去重窗口
	})

	fmt.Println("📝 使用 Nats-Msg-Id 实现幂等发布：")

	// 发布带去重 ID 的消息
	msg := &nats.Msg{
		Subject: "orders.handson.created",
		Header: nats.Header{
			"Nats-Msg-Id": []string{"unique-order-001"},
		},
		Data: []byte(`{"order_id":"ORD-001","amount":100}`),
	}

	// 第一次发布
	ack, err := js.PublishMsg(ctx, msg)
	if err != nil {
		fmt.Printf("  ❌ 发布失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 首次发布: seq=%d duplicate=%v\n", ack.Sequence, ack.Duplicate)
	}

	// 重复发布（相同 Msg-Id）
	ack, err = js.PublishMsg(ctx, msg)
	if err != nil {
		fmt.Printf("  ❌ 发布失败: %v\n", err)
	} else {
		fmt.Printf("  ⚠️ 重复发布: seq=%d duplicate=%v（已去重）\n", ack.Sequence, ack.Duplicate)
	}
	fmt.Println()

	// ========================================
	// 第六部分：实战练习
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第六部分：实战练习")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("练习：实现一个可靠的消息处理流程")
	fmt.Println()
	fmt.Println("需求：")
	fmt.Println("  1. 发布订单事件到 Stream")
	fmt.Println("  2. Consumer 消费并处理")
	fmt.Println("  3. 处理成功 ACK，失败 NAK 重试")
	fmt.Println()
	fmt.Print("按 Enter 键运行示例...")
	// fmt.Scanln() // 自动运行模式，跳过等待

	// 发布新消息
	fmt.Println("\n📝 发布测试消息：")
	for i := 4; i <= 6; i++ {
		order := OrderEvent{
			OrderID: fmt.Sprintf("ORD-%03d", i),
			Amount:  float64(i * 100),
			Status:  "created",
		}
		data, _ := json.Marshal(order)
		ack, _ := js.Publish(ctx, "orders.handson.created", data)
		fmt.Printf("  ✅ 发布: seq=%d\n", ack.Sequence)
	}

	// 消费处理
	fmt.Println("\n📝 消费处理：")
	msgs, _ = consumer.Fetch(10, jetstream.FetchMaxWait(2*time.Second))
	for msg := range msgs.Messages() {
		var order OrderEvent
		json.Unmarshal(msg.Data(), &order)

		meta, _ := msg.Metadata()

		// 模拟处理（偶发失败）
		if order.OrderID == "ORD-005" {
			fmt.Printf("  ❌ 处理失败: %s (NAK 重试)\n", order.OrderID)
			msg.Nak()
		} else {
			_ = meta // 使用 meta 避免编译警告
			fmt.Printf("  ✅ 处理成功: %s (ACK)\n", order.OrderID)
			msg.Ack()
		}
	}

	// 再次拉取，查看 NAK 的消息是否重投
	time.Sleep(100 * time.Millisecond)
	fmt.Println("\n📝 再次拉取（NAK 的消息会重投）：")
	msgs, _ = consumer.Fetch(10, jetstream.FetchMaxWait(2*time.Second))
	for msg := range msgs.Messages() {
		var order OrderEvent
		json.Unmarshal(msg.Data(), &order)
		meta, _ := msg.Metadata()

		fmt.Printf("  📨 重投消息: %s (第 %d 次投递)\n", order.OrderID, meta.NumDelivered)
		msg.Ack()
	}

	// 清理
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第四章学习完成！")
	fmt.Println("下一篇：05-kv - KV Store 状态管理")
	fmt.Println("════════════════════════════════════════════════════════════")

	// 删除测试 Stream（可选）
	// js.DeleteStream(ctx, "ORDERS_HANDSON")
}