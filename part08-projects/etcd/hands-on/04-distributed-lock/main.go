package main

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
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

	fmt.Println("=== 04-distributed-lock: 分布式锁演示 ===")

	// 模拟两个客户端竞争锁
	fmt.Println("\n模拟两个客户端竞争分布式锁...")

	go clientWithLock(cli, "Client-A", 1)
	go clientWithLock(cli, "Client-B", 2)

	// 等待所有客户端完成
	time.Sleep(10 * time.Second)
	fmt.Println("\n演示结束")
}

func clientWithLock(cli *clientv3.Client, name string, delay int) {
	fmt.Printf("%s: 开始尝试获取锁...\n", name)

	// 创建 Session
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		fmt.Printf("%s: 创建 Session 失败: %v\n", name, err)
		return
	}
	defer session.Close()

	// 创建 Mutex
	mutex := concurrency.NewMutex(session, "/locks/my-resource")

	// 尝试获取锁
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	err = mutex.Lock(ctx)
	if err != nil {
		fmt.Printf("%s: 获取锁失败: %v\n", name, err)
		return
	}

	waitTime := time.Since(start)
	fmt.Printf("%s: 获取锁成功（等待 %.2f 秒）\n", name, waitTime.Seconds())

	// 执行关键工作
	fmt.Printf("%s: 执行关键工作...\n", name)
	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Printf("%s: 工作完成\n", name)

	// 释放锁
	err = mutex.Unlock(context.Background())
	if err != nil {
		fmt.Printf("%s: 释放锁失败: %v\n", name, err)
		return
	}
	fmt.Printf("%s: 释放锁成功\n", name)
}
