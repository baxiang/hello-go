// 02 - Pub/Sub 发布订阅实践
//
// 理论知识：参见 ../03-pub-sub.md
//
// 核心概念：
// - 发布者发送消息到 Subject，不关心谁接收
// - 订阅者监听 Subject，收到所有匹配的消息
// - 扇出（Fan-out）：一条消息同时投递给所有订阅者
// - At Most Once：消息最多投递一次，不保证送达
//
// 运行方式：
//   go run ./02-pub-sub/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const natsURL = "nats://localhost:4222"

// 订单事件
type OrderEvent struct {
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	nc, err := nats.Connect(natsURL,
		nats.Name("hands-on-02-pubsub"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Drain()

	fmt.Println("✅ 已连接到 NATS:", nc.ConnectedUrl())
	fmt.Println()

	// ========================================
	// 第一部分：基础 Pub/Sub
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第一部分：基础 Pub/Sub")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// 异步订阅（推荐）
	fmt.Println("📝 异步订阅示例：")
	sub, _ := nc.Subscribe("demo.hello", func(msg *nats.Msg) {
		fmt.Printf("  📨 收到消息: subject=%s data=%s\n", msg.Subject, string(msg.Data))
	})
	nc.Flush()

	// 发布消息
	nc.Publish("demo.hello", []byte("Hello NATS!"))
	nc.Publish("demo.hello", []byte("Hello World!"))
	time.Sleep(100 * time.Millisecond)

	sub.Unsubscribe()
	fmt.Println()

	// ========================================
	// 第二部分：扇出（Fan-out）机制
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第二部分：扇出（Fan-out）机制")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("扇出：一条消息同时投递给所有订阅者")
	fmt.Println()

	var wg sync.WaitGroup

	// 模拟三个独立服务订阅同一个 Subject
	fmt.Println("启动 3 个订阅者（模拟库存服务、通知服务、分析服务）：")

	// 库存服务
	wg.Add(1)
	nc.Subscribe("orders.created", func(msg *nats.Msg) {
		var order OrderEvent
		json.Unmarshal(msg.Data, &order)
		fmt.Printf("  📦 [库存服务] 锁定库存: order_id=%s\n", order.OrderID)
		time.Sleep(10 * time.Millisecond) // 模拟处理
		wg.Done()
	})

	// 通知服务
	wg.Add(1)
	nc.Subscribe("orders.created", func(msg *nats.Msg) {
		var order OrderEvent
		json.Unmarshal(msg.Data, &order)
		fmt.Printf("  📧 [通知服务] 发送确认邮件: user_id=%s\n", order.UserID)
		time.Sleep(15 * time.Millisecond) // 模拟处理
		wg.Done()
	})

	// 分析服务
	wg.Add(1)
	nc.Subscribe("orders.created", func(msg *nats.Msg) {
		var order OrderEvent
		json.Unmarshal(msg.Data, &order)
		fmt.Printf("  📊 [分析服务] 记录订单: amount=%.2f\n", order.Amount)
		time.Sleep(5 * time.Millisecond) // 模拟处理
		wg.Done()
	})

	nc.Flush()

	// 发布一条消息，3 个服务同时收到
	order := OrderEvent{
		OrderID:   "ORD-001",
		UserID:    "U-123",
		Amount:    299.00,
		Status:    "created",
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(order)

	fmt.Printf("\n📤 发布订单事件: %s\n", string(data))
	nc.Publish("orders.created", data)

	wg.Wait()
	fmt.Println("\n✅ 所有服务处理完成（扇出完成）")
	fmt.Println()

	// ========================================
	// 第三部分：同步订阅 vs 异步订阅
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第三部分：同步订阅 vs 异步订阅")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("对比：")
	fmt.Println("┌────────────┬────────────────────────┬────────────────────────┐")
	fmt.Println("│ 特性       │ 异步订阅               │ 同步订阅               │")
	fmt.Println("├────────────┼────────────────────────┼────────────────────────┤")
	fmt.Println("│ 使用场景   │ 持续监听，生产环境主流  │ 测试、脚本、精确控制   │")
	fmt.Println("│ 阻塞行为   │ 不阻塞，回调触发       │ 阻塞等待               │")
	fmt.Println("│ 并发处理   │ 默认串行               │ 完全由调用方控制       │")
	fmt.Println("└────────────┴────────────────────────┴────────────────────────┘")
	fmt.Println()

	// 同步订阅示例
	fmt.Println("📝 同步订阅示例：")
	syncSub, _ := nc.SubscribeSync("demo.sync")
	nc.Flush()

	// 另一个 goroutine 发布消息
	go func() {
		time.Sleep(500 * time.Millisecond)
		nc.Publish("demo.sync", []byte("同步消息测试"))
	}()

	fmt.Println("  等待消息（最多 2 秒）...")
	msg, err := syncSub.NextMsg(2 * time.Second)
	if err != nil {
		fmt.Printf("  ❌ 超时或错误: %v\n", err)
	} else {
		fmt.Printf("  📨 收到: %s\n", string(msg.Data))
	}
	syncSub.Unsubscribe()
	fmt.Println()

	// ========================================
	// 第四部分：消息 Headers（NATS 2.2+）
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第四部分：消息 Headers")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// 订阅带 Headers 的消息
	nc.Subscribe("demo.headers", func(msg *nats.Msg) {
		fmt.Printf("  📨 收到消息:\n")
		fmt.Printf("     Subject: %s\n", msg.Subject)
		fmt.Printf("     Headers:\n")
		for key, values := range msg.Header {
			fmt.Printf("       %s: %s\n", key, values[0])
		}
		fmt.Printf("     Payload: %s\n", string(msg.Data))
	})
	nc.Flush()

	// 发布带 Headers 的消息
	headerMsg := &nats.Msg{
		Subject: "demo.headers",
		Header: nats.Header{
			"Content-Type":   []string{"application/json"},
			"X-Trace-Id":     []string{"trace-abc-123"},
			"X-Source-Service": []string{"order-service"},
		},
		Data: []byte(`{"event":"test"}`),
	}
	nc.PublishMsg(headerMsg)
	time.Sleep(100 * time.Millisecond)
	fmt.Println()

	// ========================================
	// 第五部分：Slow Consumer 问题
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第五部分：Slow Consumer 问题")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("Slow Consumer：消息消费速度低于生产速度")
	fmt.Println()
	fmt.Println("问题表现：")
	fmt.Println("  1. 服务器端缓冲区积压")
	fmt.Println("  2. 消息开始丢失")
	fmt.Println("  3. 服务器发送 -ERR 'Slow Consumer'")
	fmt.Println()
	fmt.Println("解决方案：")
	fmt.Println("  1. 使用工作池提高处理速度")
	fmt.Println("  2. 使用 Queue Group 分散压力")
	fmt.Println("  3. 使用 JetStream 实现背压")
	fmt.Println()

	// 工作池示例
	fmt.Println("📝 工作池模式示例：")

	jobs := make(chan *nats.Msg, 100)
	numWorkers := 3

	// 启动工作池
	for i := 1; i <= numWorkers; i++ {
		go func(id int) {
			for msg := range jobs {
				time.Sleep(10 * time.Millisecond) // 模拟处理
				fmt.Printf("  👷 [Worker %d] 处理: %s\n", id, string(msg.Data))
			}
		}(i)
	}

	// 订阅者只做非阻塞入队
	nc.Subscribe("demo.workpool", func(msg *nats.Msg) {
		select {
		case jobs <- msg:
			// 成功入队
		default:
			fmt.Println("  ⚠️ 队列已满，丢弃消息")
		}
	})
	nc.Flush()

	// 快速发布 10 条消息
	fmt.Println("  发布 10 条消息...")
	for i := 0; i < 10; i++ {
		nc.Publish("demo.workpool", []byte(fmt.Sprintf("msg-%d", i+1)))
	}

	time.Sleep(500 * time.Millisecond)
	close(jobs)
	fmt.Println()

	// ========================================
	// 第六部分：实战练习
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第六部分：实战练习")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("练习：实现一个简单的事件广播系统")
	fmt.Println()
	fmt.Println("需求：")
	fmt.Println("  1. 发布者发送用户注册事件")
	fmt.Println("  2. 积分服务监听并发放欢迎积分")
	fmt.Println("  3. 通知服务监听并发送欢迎邮件")
	fmt.Println("  4. 日志服务监听并记录注册日志")
	fmt.Println()
	fmt.Print("按 Enter 键运行示例...")
	// fmt.Scanln() // 自动运行模式，跳过等待

	// 清理之前的订阅
	cleanupSub, _ := nc.Subscribe("user.registered", nil)
	cleanupSub.Unsubscribe()

	var exerciseWg sync.WaitGroup

	// 积分服务
	exerciseWg.Add(1)
	nc.Subscribe("user.registered", func(msg *nats.Msg) {
		fmt.Printf("  🎁 [积分服务] 发放欢迎积分: %s\n", string(msg.Data))
		exerciseWg.Done()
	})

	// 通知服务
	exerciseWg.Add(1)
	nc.Subscribe("user.registered", func(msg *nats.Msg) {
		fmt.Printf("  📧 [通知服务] 发送欢迎邮件: %s\n", string(msg.Data))
		exerciseWg.Done()
	})

	// 日志服务
	exerciseWg.Add(1)
	nc.Subscribe("user.registered", func(msg *nats.Msg) {
		fmt.Printf("  📝 [日志服务] 记录注册日志: %s\n", string(msg.Data))
		exerciseWg.Done()
	})

	nc.Flush()

	// 发布用户注册事件
	fmt.Println("\n📤 发布用户注册事件: user.registered")
	nc.Publish("user.registered", []byte(`{"user_id":"U-456","email":"test@example.com"}`))

	exerciseWg.Wait()
	fmt.Println("\n✅ 所有服务处理完成")

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第二章学习完成！")
	fmt.Println("下一篇：03-request-reply - Request/Reply 模式")
	fmt.Println("════════════════════════════════════════════════════════════")
}