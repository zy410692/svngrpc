package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	pb "github.com/zy410692/svngrpc/server/pb"

	"google.golang.org/grpc"
)

const (
	port      = ":50051"
	authzFile = "/home/app/svnconfig/authz"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	mu sync.RWMutex
}

func (s *server) AddOrUpdatePermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 读取文件内容
	content, err := os.ReadFile(authzFile)
	if err != nil {
		return &pb.PermissionResponse{
			Success: false,
			Message: fmt.Sprintf("无法读取文件: %v", err),
		}, nil
	}

	lines := strings.Split(string(content), "\n")
	urlSection := fmt.Sprintf("[%s]", req.Url)
	urlFound := false
	userFound := false

	// 遍历文件查找URL和用户
	for i, line := range lines {
		if strings.TrimSpace(line) == urlSection {
			urlFound = true
			// 检查下一行是否包含用户
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(lines[j], "[") {
					break
				}
				if strings.HasPrefix(lines[j], req.User+"=") {
					lines[j] = fmt.Sprintf("%s=%s", req.User, req.Permissions)
					userFound = true
					break
				}
			}
			if !userFound {
				// 在URL部分添加新用户
				lines = append(lines[:i+1], append([]string{fmt.Sprintf("%s=%s", req.User, req.Permissions)}, lines[i+1:]...)...)
			}
			break
		}
	}

	if !urlFound {
		// 添加新的URL部分
		lines = append(lines, urlSection)
		lines = append(lines, fmt.Sprintf("%s=%s", req.User, req.Permissions))
	}

	// 写入文件
	err = os.WriteFile(authzFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return &pb.PermissionResponse{
			Success: false,
			Message: fmt.Sprintf("写入文件失败: %v", err),
		}, nil
	}

	return &pb.PermissionResponse{
		Success: true,
		Message: "权限更新成功",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{})
	fmt.Printf("服务器启动在 %s\n", port)
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v\n", err)
	}
}
