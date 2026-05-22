# 流式 RPC

gRPC 支持三种流式传输模式。

## 服务端流

服务端返回流式数据。

```protobuf
service StreamService {
  rpc GetStream (Request) returns (stream Response) {}
}
```

```go
func (s *server) GetStream(req *Request, stream pb.StreamService_GetStreamServer) error {
    for i := 0; i < 10; i++ {
        stream.Send(&Response{Value: int32(i)})
    }
    return nil
}
```

## 客户端流

客户端发送流式数据。

```protobuf
service StreamService {
  rpc SendStream (stream Request) returns (Response) {}
}
```

```go
func (s *server) SendStream(stream pb.StreamService_SendStreamServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&Response{Count: count})
        }
        count++
    }
}
```

## 双向流

双向同时发送流式数据。

```protobuf
service StreamService {
  rpc BidirectionalStream (stream Request) returns (stream Response) {}
}
```

## 使用场景

- 服务端流：实时推送、大文件下载
- 客户端流：大文件上传、批量请求
- 双向流：实时聊天、游戏

参考：https://grpc.io/docs/
