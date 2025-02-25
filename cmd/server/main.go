package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net"

	"github.com/socket-communication/internal/config"
	"github.com/socket-communication/internal/logger"
	"github.com/socket-communication/internal/models"
	"github.com/socket-communication/internal/server"
)

func main() {
	// 加載配置
	cfg := config.LoadConfig()

	// 初始化日誌
	if err := logger.InitLogger(cfg.LogFile); err != nil {
		log.Fatal(err)
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
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}(listener)

	log.Printf("Server started on %s", cfg.ServerPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn, cfg.BasePath)
	}
}

// 處理客戶端連接及請求
func handleConnection(conn net.Conn, basePath string) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}(conn)
	log.Printf("New connection from %s", conn.RemoteAddr())

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req models.Request
		if err := decoder.Decode(&req); err != nil {
			log.Printf("Error decoding request: %v", err)
			break // 解碼失敗，退出循環關閉連接
		}

		log.Printf("Received request: %+v", req)

		// 檢查是否為斷開連接
		if req.Command == "disconnect" {
			log.Printf("Client requested to close the connection")
			break
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
				response = server.ProcessFileRequest(req, basePath)
			}
		}

		// 發送響應
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			break // 編碼失敗，退出循環關閉連接
		}

		log.Printf("Sent response: %+v", response)
	}
}
