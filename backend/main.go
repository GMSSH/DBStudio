package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sqlmanager/handler/rpc"
	"sqlmanager/pkg/db"
	"syscall"
	"time"

	"github.com/DemonZack/simplejrpc-go/core"
)

func main() {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}
	
	// Create temp directory relative to working directory
	tempDir := filepath.Join(workDir, "..", "..", "tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	
	// Use absolute socket path
	sock, _ := filepath.Abs(filepath.Join(tempDir, "app.sock"))

	// 初始化服务

	// 启动空闲连接清理 (每 5 分钟检查，清理 15 分钟未使用的连接)
	manager := db.GetManager()
	manager.StartCleanupRoutine(5*time.Minute, 15*time.Minute)

	// 创建 RPC Server 并注册服务
	server := rpc.NewServer()

	// 监听退出信号，实现优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		core.Container.Log().Info("收到退出信号，开始优雅关闭...")

		// 关闭所有数据库连接
		manager.Shutdown()
		core.Container.Log().Info("所有数据库连接已关闭")

		os.Exit(0)
	}()

	// 启动服务 (阻塞)
	log.Printf("DBStudio starting, socket: %s", sock)
	if err := server.Start(sock); err != nil {
		core.Container.Log().Error(fmt.Sprintf("启动RPC服务器失败: %v", err))
		return
	}
}
