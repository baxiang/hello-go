package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	fmt.Println("=== 03-lease-service: 基于租约的服务注册 ===")

	ctx := context.Background()

	// 创建租约（60 秒）
	fmt.Println("\n1. 创建租约")
	leaseResp, err := cli.Lease.Grant(ctx, 60)
	if err != nil {
		panic(err)
	}
	leaseID := leaseResp.ID
	fmt.Printf("租约创建成功, ID: %d, TTL: %d 秒\n", leaseID, leaseResp.TTL)

	// 注册服务（绑定租约）
	fmt.Println("\n2. 注册服务")
	key := "/services/myapp/instance-1"
	value := "localhost:8080"
	_, err = cli.Put(ctx, key, value, clientv3.WithLease(leaseID))
	if err != nil {
		panic(err)
	}
	fmt.Printf("服务注册成功: %s -> %s\n", key, value)

	// 启动 KeepAlive
	fmt.Println("\n3. 启动 KeepAlive")
	keepAliveCh, err := cli.Lease.KeepAlive(ctx, leaseID)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case resp := <-keepAliveCh:
				if resp == nil {
					fmt.Println("KeepAlive channel closed")
					return
				}
				fmt.Printf("KeepAlive 成功, TTL: %d\n", resp.TTL)
			}
		}
	}()

	// 查询租约状态
	fmt.Println("\n4. 查询租约状态")
	ttlResp, err := cli.Lease.TimeToLive(ctx, leaseID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("租约剩余时间: %d 秒, 关联键数: %d\n", ttlResp.TTL, len(ttlResp.Keys))
	for _, k := range ttlResp.Keys {
		fmt.Printf("  关联键: %s\n", k)
	}

	// 查询服务列表
	fmt.Println("\n5. 查询服务列表")
	getResp, err := cli.Get(ctx, "/services/myapp/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	fmt.Printf("找到 %d 个服务实例\n", getResp.Count)
	for _, kv := range getResp.Kvs {
		fmt.Printf("  %s -> %s\n", kv.Key, kv.Value)
	}

	// 等待中断信号
	fmt.Println("\n6. 程序运行中，按 Ctrl+C 退出")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 撤销租约
	fmt.Println("\n7. 撤销租约")
	_, err = cli.Lease.Revoke(ctx, leaseID)
	if err != nil {
		panic(err)
	}
	fmt.Println("租约已撤销，服务自动注销")

	// 验证服务已注销
	getResp, err = cli.Get(ctx, "/services/myapp/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	fmt.Printf("服务实例数: %d\n", getResp.Count)
}
