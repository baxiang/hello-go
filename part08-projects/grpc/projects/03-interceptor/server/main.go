package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "03-interceptor/proto"
)

// server 实现 GreeterServer 接口
type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// loggingInterceptor 日志拦截器：记录请求和响应
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("[日志] 收到请求: %s", info.FullMethod)
	
	resp, err := handler(ctx, req)
	
	duration := time.Since(start)
	if err != nil {
		log.Printf("[日志] 请求失败: %s, 耗时: %v, 错误: %v", info.FullMethod, duration, err)
	} else {
		log.Printf("[日志] 请求成功: %s, 耗时: %v", info.FullMethod, duration)
	}
	return resp, err
}

// authInterceptor 认证拦截器：验证 Token
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 从 context 中获取 metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "缺少认证信息")
	}

	// 获取 token
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "缺少 Token")
	}

	// 验证 token（实际项目中应该验证 JWT 或调用认证服务）
	if tokens[0] != "Bearer valid-token" {
		return nil, status.Error(codes.Unauthenticated, "无效的 Token")
	}

	log.Printf("[认证] 用户认证通过")
	return handler(ctx, req)
}

// recoveryInterceptor 恢复拦截器：捕获 panic
func recoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[恢复] 捕获 panic: %v", r)
			err = status.Error(codes.Internal, fmt.Sprintf("内部错误: %v", r))
		}
	}()
	return handler(ctx, req)
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	// 注册多个拦截器（按顺序执行）
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor,  // 最先执行，最后返回
			loggingInterceptor,   // 中间执行
			authInterceptor,      // 最后执行，最先返回
		),
	)
	pb.RegisterGreeterServer(s, &server{})

	log.Println("拦截器服务启动，监听端口 :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}