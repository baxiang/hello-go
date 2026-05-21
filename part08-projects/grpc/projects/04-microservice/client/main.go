package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "04-microservice/proto"
)

func main() {
	var creds credentials.TransportCredentials
	var err error

	// 尝试加载 CA 证书进行 TLS 连接
	creds, err = credentials.NewClientTLSFromFile("ca.crt", "localhost")
	if err != nil {
		log.Printf("加载 CA 证书失败，使用不安全连接: %v", err)
		creds = insecure.NewCredentials()
	}

	// 连接单个服务（实际生产环境应使用服务发现和负载均衡）
	conn, err := grpc.Dial("localhost:50054",
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 发送多个请求，观察负载均衡效果
	for i := 0; i < 5; i++ {
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: fmt.Sprintf("Client-%d", i)})
		if err != nil {
			log.Printf("请求失败: %v", err)
			continue
		}
		fmt.Printf("收到响应: %s (来自 %s)\n", resp.GetMessage(), resp.GetServerId())
		time.Sleep(500 * time.Millisecond)
	}

	// 获取服务统计信息
	stats, err := client.GetStats(ctx, &pb.StatsRequest{})
	if err != nil {
		log.Printf("获取统计失败: %v", err)
		return
	}
	fmt.Printf("\n服务统计: 请求数=%d, 活跃连接=%d\n", stats.GetRequestCount(), stats.GetActiveConnections())
}