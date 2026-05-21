package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "01-helloworld/proto"
)

func main() {
	// 建立连接
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := pb.NewGreeterClient(conn)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 调用服务
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	fmt.Printf("收到服务端响应: %s\n", resp.GetMessage())
}