package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "01-helloworld/proto"
)

// server 实现了 GreeterServer 接口
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello 处理客户端请求
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("收到客户端请求: %s", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	// 监听 TCP 端口
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	// 创建 gRPC 服务
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	log.Println("gRPC 服务启动，监听端口 :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}