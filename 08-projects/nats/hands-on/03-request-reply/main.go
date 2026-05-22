// 03 - Request/Reply 实践
//
// 理论知识：参见 ../04-request-reply.md
//
// 核心概念：
// - Request/Reply 是 NATS 内置的 RPC 模式
// - 请求者发送消息并等待响应
// - 响应者订阅 Subject 并回复
// - 使用 _INBOX 作为临时回复地址
// - Queue Group 实现负载均衡
//
// 运行方式：
//   go run ./03-request-reply/main.go
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

// 用户请求/响应
type UserRequest struct {
	UserID string `json:"user_id"`
}

type UserResponse struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Error  string `json:"error,omitempty"`
}

func main() {
	nc, err := nats.Connect(natsURL,
		nats.Name("hands-on-03-rpc"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Drain()

	fmt.Println("✅ 已连接到 NATS:", nc.ConnectedUrl())
	fmt.Println()

	// ========================================
	// 第一部分：基础 Request/Reply
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第一部分：基础 Request/Reply")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("Request/Reply 流程：")
	fmt.Println()
	fmt.Println("  请求者                              响应者")
	fmt.Println("    │                                   │")
	fmt.Println("    │  订阅 _INBOX.xxx（自动生成）       │")
	fmt.Println("    │──────────────────────────────────▶│")
	fmt.Println("    │                                   │")
	fmt.Println("    │  PUB rpc.user.get                 │")
	fmt.Println("    │  Reply-To: _INBOX.xxx             │")
	fmt.Println("    │──────────────────────────────────▶│")
	fmt.Println("    │                                   │ 处理请求")
	fmt.Println("    │                                   │")
	fmt.Println("    │  MSG _INBOX.xxx（响应）           │")
	fmt.Println("    │◀──────────────────────────────────│")
	fmt.Println("    │                                   │")
	fmt.Println()

	// 启动响应者
	fmt.Println("📝 启动用户服务（响应者）：")
	nc.Subscribe("rpc.user.get", func(msg *nats.Msg) {
		var req UserRequest
		json.Unmarshal(msg.Data, &req)

		fmt.Printf("  📨 [用户服务] 收到请求: user_id=%s\n", req.UserID)

		// 模拟数据库查询
		resp := UserResponse{
			UserID: req.UserID,
			Name:   fmt.Sprintf("用户_%s", req.UserID),
			Email:  fmt.Sprintf("%s@example.com", req.UserID),
		}

		data, _ := json.Marshal(resp)
		msg.Respond(data)
		fmt.Printf("  ✅ [用户服务] 已响应\n")
	})
	nc.Flush()

	// 请求者发送请求
	fmt.Println("\n📝 请求者发送请求：")
	reqData, _ := json.Marshal(UserRequest{UserID: "U-001"})

	resp, err := nc.Request("rpc.user.get", reqData, 2*time.Second)
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
	} else {
		var userResp UserResponse
		json.Unmarshal(resp.Data, &userResp)
		fmt.Printf("  📨 收到响应: %+v\n", userResp)
	}
	fmt.Println()

	// ========================================
	// 第二部分：超时处理
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第二部分：超时处理")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// 慢响应服务
	nc.Subscribe("rpc.slow", func(msg *nats.Msg) {
		fmt.Println("  ⏳ [慢服务] 开始处理（需要 2 秒）...")
		time.Sleep(2 * time.Second)
		msg.Respond([]byte("完成"))
	})
	nc.Flush()

	// 1 秒超时
	fmt.Println("📝 发送请求（超时 1 秒）：")
	_, err = nc.Request("rpc.slow", []byte("test"), 1*time.Second)
	if err != nil {
		if err == nats.ErrTimeout {
			fmt.Println("  ⚠️ 请求超时（1 秒内未收到响应）")
		} else {
			fmt.Printf("  ❌ 错误: %v\n", err)
		}
	}
	fmt.Println()

	// ========================================
	// 第三部分：无响应者处理
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第三部分：无响应者处理")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("📝 请求不存在的服务：")
	_, err = nc.Request("rpc.nonexistent", []byte("test"), 2*time.Second)
	if err != nil {
		if err == nats.ErrNoResponders {
			fmt.Println("  ⚠️ 没有响应者（服务未启动或不存在）")
		} else {
			fmt.Printf("  ❌ 错误: %v\n", err)
		}
	}
	fmt.Println()

	// ========================================
	// 第四部分：Queue Group 负载均衡
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第四部分：Queue Group 负载均衡")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("Queue Group：多个实例共享同一个队列，每条消息只投递给一个实例")
	fmt.Println()

	// 启动 3 个工作实例
	fmt.Println("启动 3 个工作实例（共享 queue group 'workers'）：")
	for i := 1; i <= 3; i++ {
		workerID := i
		nc.QueueSubscribe("rpc.task", "workers", func(msg *nats.Msg) {
			fmt.Printf("  👷 [Worker %d] 处理任务: %s\n", workerID, string(msg.Data))
			time.Sleep(50 * time.Millisecond)
			msg.Respond([]byte(fmt.Sprintf("Worker %d 完成", workerID)))
		})
	}
	nc.Flush()

	// 发送 6 个请求，会被均匀分配
	fmt.Println("\n发送 6 个请求：")
	for i := 1; i <= 6; i++ {
		resp, err := nc.Request("rpc.task", []byte(fmt.Sprintf("task-%d", i)), 2*time.Second)
		if err != nil {
			fmt.Printf("  ❌ task-%d 失败: %v\n", i, err)
		} else {
			fmt.Printf("  📨 task-%d 响应: %s\n", i, string(resp.Data))
		}
	}
	fmt.Println()

	// ========================================
	// 第五部分：与 HTTP 对比
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第五部分：NATS RPC vs HTTP")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("对比：")
	fmt.Println("┌────────────────┬─────────────────────┬─────────────────────┐")
	fmt.Println("│ 特性           │ NATS RPC            │ HTTP                │")
	fmt.Println("├────────────────┼─────────────────────┼─────────────────────┤")
	fmt.Println("│ 延迟           │ < 1ms               │ 10-100ms            │")
	fmt.Println("│ 连接开销       │ 长连接复用          │ 每次请求新建连接    │")
	fmt.Println("│ 服务发现       │ 自动（订阅即注册）  │ 需要额外组件        │")
	fmt.Println("│ 负载均衡       │ Queue Group 内置    │ 需要负载均衡器      │")
	fmt.Println("│ 超时控制       │ 客户端控制          │ 客户端/服务端控制   │")
	fmt.Println("└────────────────┴─────────────────────┴─────────────────────┘")
	fmt.Println()

	// ========================================
	// 第六部分：实战练习
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第六部分：实战练习")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("练习：实现一个简单的微服务调用")
	fmt.Println()
	fmt.Println("需求：")
	fmt.Println("  1. 订单服务调用库存服务检查库存")
	fmt.Println("  2. 库存服务返回是否有库存")
	fmt.Println("  3. 处理超时和无响应情况")
	fmt.Println()
	fmt.Print("按 Enter 键运行示例...")
	// fmt.Scanln() // 自动运行模式，跳过等待

	// 库存服务
	nc.QueueSubscribe("rpc.inventory.check", "inventory-service", func(msg *nats.Msg) {
		type InventoryReq struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}
		type InventoryResp struct {
			Available bool   `json:"available"`
			Stock     int    `json:"stock"`
		}

		var req InventoryReq
		json.Unmarshal(msg.Data, &req)

		fmt.Printf("  📦 [库存服务] 检查库存: product=%s qty=%d\n", req.ProductID, req.Quantity)

		// 模拟库存检查
		stock := 100
		resp := InventoryResp{
			Available: stock >= req.Quantity,
			Stock:     stock,
		}

		data, _ := json.Marshal(resp)
		msg.Respond(data)
	})
	nc.Flush()

	// 订单服务调用库存服务
	fmt.Println("\n📝 订单服务调用库存服务：")

	req := struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}{
		ProductID: "PROD-001",
		Quantity:  5,
	}
	reqData, _ = json.Marshal(req)

	resp, err = nc.Request("rpc.inventory.check", reqData, 2*time.Second)
	if err != nil {
		if err == nats.ErrNoResponders {
			fmt.Println("  ❌ 库存服务不可用")
		} else if err == nats.ErrTimeout {
			fmt.Println("  ❌ 库存服务响应超时")
		} else {
			fmt.Printf("  ❌ 错误: %v\n", err)
		}
	} else {
		type InventoryResp struct {
			Available bool `json:"available"`
			Stock     int  `json:"stock"`
		}
		var invResp InventoryResp
		json.Unmarshal(resp.Data, &invResp)

		if invResp.Available {
			fmt.Printf("  ✅ 库存充足: stock=%d\n", invResp.Stock)
		} else {
			fmt.Println("  ❌ 库存不足")
		}
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第三章学习完成！")
	fmt.Println("下一篇：04-jetstream - JetStream 持久化")
	fmt.Println("════════════════════════════════════════════════════════════")
}

// 避免未使用变量警告
var _ = sync.Mutex{}