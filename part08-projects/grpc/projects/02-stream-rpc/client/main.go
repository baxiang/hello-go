package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "02-stream-rpc/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewStreamServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 服务端流
	fmt.Println("=== 服务端流 ===")
	stream1, err := client.ServerStream(ctx, &pb.StreamRequest{Data: "Hello"})
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}
	for {
		resp, err := stream1.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("接收失败: %v", err)
		}
		fmt.Printf("收到: %s\n", resp.GetResult())
	}

	// 2. 客户端流
	fmt.Println("\n=== 客户端流 ===")
	stream2, err := client.ClientStream(ctx)
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := stream2.Send(&pb.StreamRequest{Data: fmt.Sprintf("消息%d", i)}); err != nil {
			log.Fatalf("发送失败: %v", err)
		}
	}
	resp, err := stream2.CloseAndRecv()
	if err != nil {
		log.Fatalf("接收失败: %v", err)
	}
	fmt.Printf("收到: %s\n", resp.GetResult())

	// 3. 双向流
	fmt.Println("\n=== 双向流 ===")
	stream3, err := client.BidirectionalStream(ctx)
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	// 发送消息
	go func() {
		for i := 0; i < 3; i++ {
			if err := stream3.Send(&pb.StreamRequest{Data: fmt.Sprintf("双向%d", i)}); err != nil {
				log.Printf("发送失败: %v", err)
				return
			}
			time.Sleep(time.Second)
		}
		stream3.CloseSend()
	}()

	// 接收消息
	for {
		resp, err := stream3.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("接收失败: %v", err)
		}
		fmt.Printf("收到: %s\n", resp.GetResult())
	}
}