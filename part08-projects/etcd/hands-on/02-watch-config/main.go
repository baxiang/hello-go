package main

import (
	"context"
	"fmt"
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

	fmt.Println("=== 02-watch-config: 监听配置变更 ===")

	ctx := context.Background()

	// 启动 Watch 监听
	fmt.Println("\n启动 Watch 监听 /config/")
	watchCh := cli.Watch(ctx, "/config/", clientv3.WithPrefix())

	// 模拟配置写入
	go func() {
		time.Sleep(1 * time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cli.Put(ctx, "/config/database/host", "localhost")
		time.Sleep(500 * time.Millisecond)

		cli.Put(ctx, "/config/database/port", "3306")
		time.Sleep(500 * time.Millisecond)

		cli.Put(ctx, "/config/database/host", "192.168.1.100")
		time.Sleep(500 * time.Millisecond)

		cli.Delete(ctx, "/config/database/port")
	}()

	// 处理 Watch 事件
	eventCount := 0
	for watchResp := range watchCh {
		for _, event := range watchResp.Events {
			eventCount++
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Printf("事件 %d: PUT - Key: %s, Value: %s\n",
					eventCount, event.Kv.Key, event.Kv.Value)
			case clientv3.EventTypeDelete:
				fmt.Printf("事件 %d: DELETE - Key: %s\n",
					eventCount, event.Kv.Key)
			}

			// 收到 4 个事件后退出
			if eventCount >= 4 {
				fmt.Println("\n监听结束")
				return
			}
		}
	}
}
