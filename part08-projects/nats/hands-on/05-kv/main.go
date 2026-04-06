// 05 - KV Store 实践
//
// 理论知识：参见 ../07-jetstream-kv.md
//
// 核心概念：
// - KV Store 是基于 JetStream 的键值存储
// - 支持版本历史和 Watch
// - 适合存储配置、设备状态等
// - 支持 CAS（Compare-and-Swap）操作
//
// 运行方式：
//   go run ./05-kv/main.go
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

// 设备状态
type DeviceStatus struct {
	DeviceID      string `json:"device_id"`
	Status        string `json:"status"` // online, offline
	LastHeartbeat int64  `json:"last_heartbeat"`
	Version       string `json:"version"`
}

func main() {
	nc, err := nats.Connect(natsURL,
		nats.Name("hands-on-05-kv"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Drain()

	fmt.Println("✅ 已连接到 NATS:", nc.ConnectedUrl())
	fmt.Println()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("创建 JetStream 失败: %v", err)
	}

	ctx := context.Background()

	// ========================================
	// 第一部分：创建 KV Bucket
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第一部分：创建 KV Bucket")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("KV Bucket：键值存储容器")
	fmt.Println()

	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:   "DEVICE_STATUS_HANDSON",
		Description: "设备状态存储（学习示例）",
		History:  5,    // 保留 5 个历史版本
		Replicas: 1,    // 开发环境用 1
		TTL:      0,    // 0 表示永不过期
	})
	if err != nil {
		log.Fatalf("创建 KV Bucket 失败: %v", err)
	}

	fmt.Printf("✅ KV Bucket 已创建: %s\n", "DEVICE_STATUS_HANDSON")
	fmt.Println()

	// ========================================
	// 第二部分：基本 CRUD 操作
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第二部分：基本 CRUD 操作")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// Put：创建或更新
	fmt.Println("📝 Put - 创建/更新键值：")
	status1 := DeviceStatus{
		DeviceID:      "device-001",
		Status:        "online",
		LastHeartbeat: time.Now().Unix(),
		Version:       "1.0.0",
	}
	data1, _ := json.Marshal(status1)
	rev1, _ := kv.Put(ctx, "device-001", data1)
	fmt.Printf("  ✅ Put device-001: revision=%d\n", rev1)

	status2 := DeviceStatus{
		DeviceID:      "device-002",
		Status:        "online",
		LastHeartbeat: time.Now().Unix(),
		Version:       "1.0.0",
	}
	data2, _ := json.Marshal(status2)
	rev2, _ := kv.Put(ctx, "device-002", data2)
	fmt.Printf("  ✅ Put device-002: revision=%d\n", rev2)
	fmt.Println()

	// Get：读取
	fmt.Println("📝 Get - 读取键值：")
	entry, err := kv.Get(ctx, "device-001")
	if err != nil {
		fmt.Printf("  ❌ Get 失败: %v\n", err)
	} else {
		var status DeviceStatus
		json.Unmarshal(entry.Value(), &status)
		fmt.Printf("  📨 device-001: status=%s revision=%d\n", status.Status, entry.Revision())
	}
	fmt.Println()

	// Update：带版本检查的更新
	fmt.Println("📝 Update - 带版本检查更新：")
	newStatus := DeviceStatus{
		DeviceID:      "device-001",
		Status:        "busy",
		LastHeartbeat: time.Now().Unix(),
		Version:       "1.0.1",
	}
	newData, _ := json.Marshal(newStatus)
	rev3, err := kv.Update(ctx, "device-001", newData, rev1)
	if err != nil {
		fmt.Printf("  ❌ Update 失败（版本冲突）: %v\n", err)
	} else {
		fmt.Printf("  ✅ Update device-001: revision=%d\n", rev3)
	}

	// 尝试用旧版本更新（会失败）
	_, err = kv.Update(ctx, "device-001", newData, rev1)
	if err != nil {
		fmt.Printf("  ⚠️ 版本冲突: %v（预期行为）\n", err)
	}
	fmt.Println()

	// Delete：删除
	fmt.Println("📝 Delete - 删除键：")
	err = kv.Delete(ctx, "device-002")
	if err != nil {
		fmt.Printf("  ❌ Delete 失败: %v\n", err)
	} else {
		fmt.Println("  ✅ device-002 已删除")
	}

	// 验证删除
	_, err = kv.Get(ctx, "device-002")
	if err != nil {
		fmt.Printf("  📨 device-002 已不存在\n")
	}
	fmt.Println()

	// ========================================
	// 第三部分：Watch 监听变化
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第三部分：Watch 监听变化")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("Watch：实时监听键值变化")
	fmt.Println()

	// 启动 Watch
	watcher, err := kv.WatchAll(ctx)
	if err != nil {
		log.Fatalf("Watch 失败: %v", err)
	}
	defer watcher.Stop()

	fmt.Println("📝 启动 Watch，监听所有变化...")
	fmt.Println()

	// 在另一个 goroutine 中发布变化
	go func() {
		time.Sleep(500 * time.Millisecond)

		// 更新 device-001
		status := DeviceStatus{
			DeviceID:      "device-001",
			Status:        "offline",
			LastHeartbeat: time.Now().Unix(),
			Version:       "1.0.2",
		}
		data, _ := json.Marshal(status)
		kv.Put(ctx, "device-001", data)

		time.Sleep(300 * time.Millisecond)

		// 添加 device-003
		status3 := DeviceStatus{
			DeviceID:      "device-003",
			Status:        "online",
			LastHeartbeat: time.Now().Unix(),
			Version:       "1.0.0",
		}
		data3, _ := json.Marshal(status3)
		kv.Put(ctx, "device-003", data3)
	}()

	// 接收变化
	fmt.Println("接收变化（等待 2 秒）：")
	timeout := time.After(2 * time.Second)
	for {
		select {
		case entry, ok := <-watcher.Updates():
			if !ok {
				fmt.Println("  Watch 已关闭")
				goto done
			}
			if entry == nil {
				continue
			}
			var status DeviceStatus
			json.Unmarshal(entry.Value(), &status)
			fmt.Printf("  📨 变化: key=%s status=%s revision=%d\n",
				entry.Key(), status.Status, entry.Revision())

		case <-timeout:
			goto done
		}
	}
done:
	fmt.Println()

	// ========================================
	// 第四部分：列出所有键
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第四部分：列出所有键")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("📝 列出所有设备：")
	keys, err := kv.ListKeys(ctx)
	if err != nil {
		fmt.Printf("  ❌ ListKeys 失败: %v\n", err)
	} else {
		for key := range keys.Keys() {
			entry, err := kv.Get(ctx, key)
			if err != nil {
				continue
			}
			var status DeviceStatus
			json.Unmarshal(entry.Value(), &status)
			fmt.Printf("  📱 %s: status=%s\n", key, status.Status)
		}
	}
	fmt.Println()

	// ========================================
	// 第五部分：CAS 操作实现分布式锁
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第五部分：CAS 操作实现分布式锁")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("CAS（Compare-and-Swap）：原子性更新")
	fmt.Println()

	// 创建锁
	fmt.Println("📝 尝试获取锁：")
	lockKey := "lock.resource-a"

	// 第一次创建（成功）
	rev, err := kv.Create(ctx, lockKey, []byte("locked-by-client-1"))
	if err != nil {
		fmt.Printf("  ❌ 获取锁失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ client-1 获取锁成功: revision=%d\n", rev)
	}

	// 另一个客户端尝试获取（失败）
	_, err = kv.Create(ctx, lockKey, []byte("locked-by-client-2"))
	if err != nil {
		fmt.Printf("  ⚠️ client-2 获取锁失败（锁已被占用）\n")
	}
	fmt.Println()

	// 释放锁
	fmt.Println("📝 释放锁：")
	err = kv.Delete(ctx, lockKey)
	if err != nil {
		fmt.Printf("  ❌ 释放锁失败: %v\n", err)
	} else {
		fmt.Println("  ✅ 锁已释放")
	}

	// 再次尝试获取
	rev, err = kv.Create(ctx, lockKey, []byte("locked-by-client-2"))
	if err != nil {
		fmt.Printf("  ❌ 获取锁失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ client-2 获取锁成功: revision=%d\n", rev)
	}
	fmt.Println()

	// ========================================
	// 第六部分：实战练习
	// ========================================
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第六部分：实战练习")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("练习：实现设备状态管理")
	fmt.Println()
	fmt.Println("需求：")
	fmt.Println("  1. 设备上线时注册状态")
	fmt.Println("  2. 设备定期更新心跳")
	fmt.Println("  3. 客户端 Watch 监听设备状态变化")
	fmt.Println()
	fmt.Print("按 Enter 键运行示例...")
	// fmt.Scanln() // 自动运行模式，跳过等待

	// 清理之前的测试数据
	kv.Delete(ctx, "device-001")
	kv.Delete(ctx, "device-003")
	kv.Delete(ctx, lockKey)

	// 设备上线
	fmt.Println("\n📝 设备上线：")
	deviceOnline := DeviceStatus{
		DeviceID:      "device-100",
		Status:        "online",
		LastHeartbeat: time.Now().Unix(),
		Version:       "2.0.0",
	}
	data, _ := json.Marshal(deviceOnline)
	kv.Put(ctx, "device-100", data)
	fmt.Println("  ✅ device-100 已上线")

	// 模拟心跳更新
	fmt.Println("\n📝 心跳更新：")
	for i := 0; i < 3; i++ {
		time.Sleep(200 * time.Millisecond)
		deviceOnline.LastHeartbeat = time.Now().Unix()
		data, _ = json.Marshal(deviceOnline)
		rev, _ := kv.Put(ctx, "device-100", data)
		fmt.Printf("  💓 心跳更新: revision=%d\n", rev)
	}

	// 设备下线
	fmt.Println("\n📝 设备下线：")
	deviceOnline.Status = "offline"
	data, _ = json.Marshal(deviceOnline)
	kv.Put(ctx, "device-100", data)
	fmt.Println("  ✅ device-100 已下线")

	// 查看历史版本
	fmt.Println("\n📝 查看历史版本：")
	for rev := 1; rev <= 5; rev++ {
		entry, err := kv.GetRevision(ctx, "device-100", uint64(rev))
		if err != nil {
			continue
		}
		var status DeviceStatus
		json.Unmarshal(entry.Value(), &status)
		fmt.Printf("  📜 revision=%d: status=%s\n", rev, status.Status)
	}

	// 清理
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("第五章学习完成！")
	fmt.Println("恭喜你完成了 NATS 边学边练系列！")
	fmt.Println("════════════════════════════════════════════════════════════")

	// 删除测试 Bucket（可选）
	// js.DeleteKeyValue(ctx, "DEVICE_STATUS_HANDSON")
}