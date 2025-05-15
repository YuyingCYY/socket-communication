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
	"time"

	"github.com/socket-communication/internal/config"
	"github.com/socket-communication/internal/models"
	"github.com/socket-communication/internal/server"
)

type socketServer struct {
	cfg      *config.Config     // 伺服器設定
	lis      net.Listener       // 監聽器
	ctx      context.Context    // 全域 context 用於關閉協程
	cancel   context.CancelFunc // 關閉 context 的函數
	wg       sync.WaitGroup     // 等待所有連線處理結束
	sigChan  chan os.Signal     // 系統訊號通道
	doneChan chan struct{}      // 所有連線關閉通知通道
}

// 建構函數，初始化 socketServer 實例
func NewSocketServer(cfg *config.Config) *socketServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &socketServer{
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		sigChan:  make(chan os.Signal, 1),
		doneChan: make(chan struct{}),
	}
}

// 建立 TLS 設定，包含憑證與客戶端驗證
func (s *socketServer) setupTLS() (*tls.Config, error) {
	// 加載 TLS 證書
	cert, err := tls.LoadX509KeyPair(s.cfg.CertFile, s.cfg.KeyFile)
	if err != nil {
		log.Fatalf("[ERROR %d]Failed to load TLS certificates: %v", models.ErrServerTimeoutTLSLoad, err)
	}

	caCert, err := os.ReadFile(s.cfg.CertFile)
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates:          []tls.Certificate{cert},
		ClientAuth:            tls.RequireAndVerifyClientCert,
		ClientCAs:             caCertPool,
		MinVersion:            tls.VersionTLS12,
		VerifyPeerCertificate: s.createCertVerifier(),
	}, nil
}

// 回傳一個驗證客戶端憑證的函數
func (s *socketServer) createCertVerifier() func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		// 解析客戶端憑證
		clientCert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("憑證解析失敗: %v", err)
		}

		// 檢查指紋白名單，使用 %X 強制大寫
		fingerprintHex := fmt.Sprintf("%X", sha256.Sum256(clientCert.Raw))
		if fingerprintHex != s.cfg.AllowedClient {
			log.Printf("非法憑證指紋: %s (CN=%s, SN=%s)",
				fingerprintHex,
				clientCert.Subject.CommonName,
				clientCert.SerialNumber)
			return fmt.Errorf("憑證未授權")
		}

		return nil
	}
}

// 處理單個客戶端連接的生命週期
func connectionHandler(ctx context.Context, conn net.Conn, basePath string, doneChan chan struct{}) {
	// 類型斷言為 TLS 連接
	tlsConn, ok := conn.(*tls.Conn)
	if ok {
		// 握手驗證
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("TLS 握手失敗: %v", err)
			conn.Close()
			doneChan <- struct{}{}
			return
		}
	}

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
		return server.ProcessRequest(req, basePath, ".bin", server.MainAppHandler, false)
	case "ParameterTable":
		return server.ProcessRequest(req, basePath, ".data", server.ParameterTableHandler, false)
	case "SecurityZone":
		return server.ProcessRequest(req, basePath, "", server.SecurityZoneHandler, true)
	case "FaultLog":
		return models.Response{}
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

// 啟動伺服器，開始監聽與接收連線
func (s *socketServer) Start() error {
	tlsConfig, err := s.setupTLS()
	if err != nil {
		return err
	}

	listener, err := tls.Listen("tcp", s.cfg.ServerPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("[ERROR %d] Failed to start server: %v",
			models.ErrServerTimeoutStart, err)
	}
	s.lis = listener

	// 註冊系統中斷訊號
	signal.Notify(s.sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Server started on %s", s.cfg.ServerPort)

	// // 啟動接收連線的 goroutine
	go s.acceptConnections()

	select {
	case <-s.sigChan:
		log.Println("Shutting down server...")
	case <-s.ctx.Done():
	}

	return s.Shutdown()
}

// 持續接受新連線，並為每個連線啟動 goroutine
func (s *socketServer) acceptConnections() {
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			if s.ctx.Err() == nil {
				log.Printf("[ERROR %d] Error accepting connection: %v",
					models.ErrServerConnectionAccept, err)
			}
			return
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			connectionHandler(s.ctx, c, s.cfg.BasePath, s.doneChan)
		}(conn)
	}
}

// 關閉伺服器，等待所有連線結束
func (s *socketServer) Shutdown() error {
	s.cancel() // 通知所有 goroutine 結束
	s.lis.Close()

	// 等待所有連線處理完成
	go func() {
		s.wg.Wait()
		close(s.doneChan)
	}()

	// 設定最大等待時間，避免卡死
	shutdownTimeout := 30 * time.Second
	select {
	case <-s.doneChan:
		log.Println("All connections closed")
	case <-time.After(shutdownTimeout):
		log.Println("Shutdown timeout, forcing close")
	}

	log.Println("Server shutdown complete")
	return nil
}
