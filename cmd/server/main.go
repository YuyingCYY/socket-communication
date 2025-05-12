package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
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

	var station *models.Station
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var req models.Request
			if err := decoder.Decode(&req); err != nil {
				log.Printf("[ERROR %d]Error decoding request: %v", models.ErrServerJSONUnmarshal, err)
				return
			}

			// 檢查是否為斷開連接
			if req.Command == "disconnect" {
				log.Printf("Client requested to close the connection")
				return
			}

			// 檢查第一次連接站點訊息並保存
			if req.Command == "connect" {
				if isValidate, s := server.ValidateStationInfo(&req, encoder); isValidate {
					log.Printf("Invalid station information, closing connection")
					return
				} else {
					station = s
				}
			}

			req.StationType = station.StationType
			req.StationName = station.StationName
			req.StationId = station.StationId
			req.OperatorId = station.OperatorId
			if isValidate, _ := server.ValidateStationInfo(&req, encoder); isValidate {
				log.Printf("[ERROR %d]Invalid station information, closing connection", models.ErrClientInvalidStation)
				return
			}

			log.Printf("Received request: %+v", req)

			// 處理請求並生成響應
			response := processRequest(req, basePath)

			// 發送響應
			if err := encoder.Encode(response); err != nil {
				log.Printf("[ERROR %d]Error encoding response: %v", models.ErrServerJSONMarshal, err)
				return
			}

			// log.Printf("Sent response: %+v", response)
			log.Println("Sent response successfully")
		}
	}
}

// 根據請求類型處理請求並返回適當的響應
func processRequest(req models.Request, basePath string) models.Response {
	// 處理連接命令
	if req.Command == "connect" {
		return models.Response{}
	}

	// 根據 askId 處理請求
	switch req.AskId {
	case "MainApp":
		return server.ProcessMainAppRequest(req, basePath)
	case "ParameterTable":
		return server.ProcessParameterTable(req, basePath)
	case "SecurityZone":
		return server.ProcessSecurityZone(req, basePath)
	case "FaultLog":
		return server.ProcessFaultLogRequest(req, basePath)
	default:
		status := models.Status{
			ErrCode: models.ErrClientInvalidAskId,
			ErrDesc: "Invalid askId",
		}
		return models.Response{
			Status: status,
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
		log.Fatalf("[ERROR %d]Failed to load TLS certificates: %v", models.ErrServerTimeoutTLSLoad, err)
	}

	caCert, err := os.ReadFile(cfg.CertFile)
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			// 解析客戶端憑證
			clientCert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("憑證解析失敗: %v", err)
			}

			// 檢查指紋白名單，使用 %X 強制大寫
			fingerprintHex := fmt.Sprintf("%X", sha256.Sum256(clientCert.Raw))
			if fingerprintHex != cfg.AllowedClient {
				log.Printf("非法憑證指紋: %s (CN=%s, SN=%s)",
					fingerprintHex,
					clientCert.Subject.CommonName,
					clientCert.SerialNumber)
				return fmt.Errorf("憑證未授權")
			}

			return nil
		},
	}

	// 啟動服務器
	listener, err := tls.Listen("tcp", cfg.ServerPort, tlsConfig)
	if err != nil {
		log.Fatalf("[ERROR %d] Failed to start server: %v", models.ErrServerTimeoutStart, err)
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
				log.Printf("[ERROR %d] Error accepting connection: %v", models.ErrServerConnectionAccept, err)
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
