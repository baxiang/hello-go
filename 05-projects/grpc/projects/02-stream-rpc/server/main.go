package main

import (
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "02-stream-rpc/proto"
)

// server 实现了 StreamServiceServer 接口
type server struct {
	pb.UnimplementedStreamServiceServer
}

// ServerStream 服务端流：服务端向客户端发送多条消息
func (s *server) ServerStream(req *pb.StreamRequest, stream pb.StreamService_ServerStreamServer) error {
	log.Printf("收到服务端流请求: %s", req.GetData())
	for i := 0; i < 5; i++ {
		if err := stream.Send(&pb.StreamResponse{
			Result: req.GetData() + " - 消息 " + string(rune('0'+i)),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

// ClientStream 客户端流：客户端向服务端发送多条消息
func (s *server) ClientStream(stream pb.StreamService_ClientStreamServer) error {
	var count int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.StreamResponse{
				Result: "共收到 " + string(rune('0'+count)) + " 条消息",
			})
		}
		if err != nil {
			return err
		}
		log.Printf("收到客户端流消息: %s", req.GetData())
		count++
	}
}

// BidirectionalStream 双向流：双方同时发送消息
func (s *server) BidirectionalStream(stream pb.StreamService_BidirectionalStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("双向流收到: %s", req.GetData())
		if err := stream.Send(&pb.StreamResponse{
			Result: "Echo: " + req.GetData(),
		}); err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterStreamServiceServer(s, &server{})

	log.Println("流式RPC服务启动，监听端口 :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}