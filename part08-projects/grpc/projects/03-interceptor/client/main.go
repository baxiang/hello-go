package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "03-interceptor/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 1. 不带 Token 的请求（应该失败）
	fmt.Println("=== 不带 Token ===")
	_, err = client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		fmt.Printf("预期错误: %v\n", err)
	}

	// 2. 带错误 Token 的请求（应该失败）
	fmt.Println("\n=== 错误 Token ===")
	md := metadata.New(map[string]string{
		"authorization": "Bearer wrong-token",
	})
	ctx2 := metadata.NewOutgoingContext(ctx, md)
	_, err = client.SayHello(ctx2, &pb.HelloRequest{Name: "World"})
	if err != nil {
		fmt.Printf("预期错误: %v\n", err)
	}

	// 3. 带正确 Token 的请求（应该成功）
	fmt.Println("\n=== 正确 Token ===")
	md2 := metadata.New(map[string]string{
		"authorization": "Bearer valid-token",
	})
	ctx3 := metadata.NewOutgoingContext(ctx, md2)
	resp, err := client.SayHello(ctx3, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}
	fmt.Printf("收到响应: %s\n", resp.GetMessage())
}