package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	pb "04-microservice/proto"
)

var (
	port     = flag.String("port", "50054", "服务端口")
	serverID = flag.String("id", "server-1", "服务实例ID")
)

type server struct {
	pb.UnimplementedGreeterServer
	requestCount int64
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	atomic.AddInt64(&s.requestCount, 1)
	return &pb.HelloReply{
		Message:  "Hello " + in.GetName() + " from " + *serverID,
		ServerId: *serverID,
	}, nil
}

func (s *server) GetStats(ctx context.Context, in *pb.StatsRequest) (*pb.StatsReply, error) {
	return &pb.StatsReply{
		RequestCount:      atomic.LoadInt64(&s.requestCount),
		ActiveConnections: 0,
	}, nil
}

func main() {
	flag.Parse()

	// 加载 TLS 证书
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Printf("加载 TLS 证书失败，使用不安全连接: %v", err)
		creds = nil
	}

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	var opts []grpc.ServerOption
	if creds != nil {
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(s, &server{})

	// 注册健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	fmt.Printf("服务 %s 启动，监听端口 :%s\n", *serverID, *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}