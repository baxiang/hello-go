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

	fmt.Println("=== 01-kv-basics: 基础 KV 操作 ===")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Put 操作
	fmt.Println("\n1. Put 操作")
	putResp, err := cli.Put(ctx, "/demo/key1", "value1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Put 成功, Revision: %d\n", putResp.Header.Revision)

	putResp, err = cli.Put(ctx, "/demo/key2", "value2")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Put 成功, Revision: %d\n", putResp.Header.Revision)

	// Get 操作
	fmt.Println("\n2. Get 操作")
	getResp, err := cli.Get(ctx, "/demo/key1")
	if err != nil {
		panic(err)
	}
	if len(getResp.Kvs) > 0 {
		kv := getResp.Kvs[0]
		fmt.Printf("Key: %s, Value: %s\n", kv.Key, kv.Value)
		fmt.Printf("CreateRevision: %d, ModRevision: %d, Version: %d\n",
			kv.CreateRevision, kv.ModRevision, kv.Version)
	}

	// 前缀查询
	fmt.Println("\n3. 前缀查询")
	getResp, err = cli.Get(ctx, "/demo/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	fmt.Printf("找到 %d 个键\n", getResp.Count)
	for _, kv := range getResp.Kvs {
		fmt.Printf("  Key: %s, Value: %s\n", kv.Key, kv.Value)
	}

	// Delete 操作
	fmt.Println("\n4. Delete 操作")
	delResp, err := cli.Delete(ctx, "/demo/key1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("删除 %d 个键\n", delResp.Deleted)

	// 验证删除
	fmt.Println("\n5. 验证删除")
	getResp, err = cli.Get(ctx, "/demo/key1")
	if err != nil {
		panic(err)
	}
	if len(getResp.Kvs) == 0 {
		fmt.Println("键已删除")
	}

	// 清理所有 demo 键
	fmt.Println("\n6. 清理所有 demo 键")
	delResp, err = cli.Delete(ctx, "/demo/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	fmt.Printf("删除 %d 个键\n", delResp.Deleted)
}
