package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/socket-communication/internal/config"
	"github.com/socket-communication/internal/logger"
	"github.com/socket-communication/internal/models"
	"github.com/socket-communication/internal/server"
)

// 處理單個客戶端連接的生命週期
func connectionHandler(ctx context.Context, conn net.Conn, basePath string, doneChan chan struct{}) {
	defer func() {
		log.Printf("Connection from %s closed", conn.RemoteAddr())
		conn.Close()
		doneChan <- struct{}{}
	}()

	log.Printf("New connection from %s", conn.RemoteAddr())

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var req models.Request
			if err := decoder.Decode(&req); err != nil {
				log.Printf("Error decoding request: %v", err)
				return
			}

			log.Printf("Received request: %+v", req)

			// 檢查是否為斷開連接
			if req.Command == "disconnect" {
				log.Printf("Client requested to close the connection")
				return
			}

			var response models.Response

			// 檢查是否為獲取 .bin 檔
			if req.IsGetBinFile {
				response = server.ProcessFileRequest(req, basePath)
			} else {
				// 根據 askId 處理請求
				switch req.AskId {
				case "MainApp":
					response = server.ProcessMainAppRequest(req, basePath)
				default:
					response = models.Response{
						Error: "Invalid askId",
					}
				}
			}

			// 發送響應
			if err := encoder.Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
				return
			}

			log.Printf("Sent response: %+v", response)
		}
	}
}

func main() {
	// 加載配置
	cfg := config.LoadConfig()

	// 初始化日誌
	if err := logger.InitLogger(cfg.LogFile); err != nil {
		panic(err)
	}

	// 加載 TLS 證書
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// 啟動服務器
	listener, err := tls.Listen("tcp", cfg.ServerPort, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// 創建上下文和取消函數
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 監聽系統中斷信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 通道用於通知所有連接完成
	doneChan := make(chan struct{})
	var connectionWg sync.WaitGroup

	log.Printf("Server started on %s", cfg.ServerPort)

	// 接收連接的 goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				break
			}

			// 為每個連接啟動一個 goroutine
			connectionWg.Add(1)
			go func(conn net.Conn) {
				defer connectionWg.Done()
				connectionHandler(ctx, conn, cfg.BasePath, doneChan)
			}(conn)
		}
	}()

	// 等待系統中斷信號
	<-sigChan
	log.Println("Shutting down server...")

	// 取消上下文以停止接受新連接
	cancel()

	// 等待所有活動連接完成
	go func() {
		connectionWg.Wait()
		close(doneChan)
	}()

	// 等待所有連接關閉
	for range doneChan {
	}

	log.Println("Server shutdown complete")
}
