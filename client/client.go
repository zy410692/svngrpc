package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/zy410692/svngrpc/client/pb"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接服务器: %v", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 发送请求
	req := &pb.PermissionRequest{
		Url:         "http://www.baidu.com",
		User:        "zhangyi",
		Permissions: "rw",
	}

	resp, err := client.AddOrUpdatePermission(ctx, req)
	if err != nil {
		log.Fatalf("调用服务失败: %v", err)
	}

	fmt.Printf("响应: %s\n", resp.Message)
}
