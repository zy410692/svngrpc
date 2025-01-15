package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/zy410692/svngrpc/client/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address = "localhost:50051"
	// 添加认证token，需要和服务端匹配
	token = "your-secret-token"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthServiceClient(conn)

	// 创建带有认证token的上下文
	md := metadata.New(map[string]string{
		"authorization": token,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 创建测试用例结构体
	type TestCase struct {
		url         string
		user        string
		permissions string
	}

	// 生成100个测试用例
	var testCases []TestCase
	for i := 1; i <= 100; i++ {
		// 根据i的值循环分配不同的权限组合
		var user, perm string
		switch i % 3 {
		case 0:
			user = "zhangsan"
			perm = "rw"
		case 1:
			user = "lisi"
			perm = "r"
		case 2:
			user = "shenlang"
			perm = "r"
		}

		testCase := TestCase{
			url:         fmt.Sprintf("http://www.example%d.com", i),
			user:        user,
			permissions: perm,
		}
		testCases = append(testCases, testCase)
	}

	var wg sync.WaitGroup
	// 记录开始时间
	startTime := time.Now()

	// 并发执行所有测试用例
	for _, tc := range testCases {
		wg.Add(1)
		go func(tc TestCase) {
			defer wg.Done()

			// 使用带认证的ctx创建超时上下文
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			req := &pb.PermissionRequest{
				Url:         tc.url,
				User:        tc.user,
				Permissions: tc.permissions,
			}

			// 使用带认证的timeoutCtx发送请求
			resp, err := c.AddOrUpdatePermission(timeoutCtx, req)
			if err != nil {
				log.Printf("处理请求失败 - URL: %s, 用户: %s, 权限: %s, 错误: %v",
					tc.url, tc.user, tc.permissions, err)
				return
			}

			fmt.Printf("请求成功 - URL: %s, 用户: %s, 权限: %s, 响应: %s\n",
				tc.url, tc.user, tc.permissions, resp.Message)
		}(tc)
	}

	// 等待所有请求完成
	wg.Wait()

	// 计算总耗时
	duration := time.Since(startTime)
	fmt.Printf("\n所有请求已完成，总耗时: %v\n", duration)
	fmt.Printf("平均每个请求耗时: %v\n", duration/100)
}
