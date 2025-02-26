package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	conn       *tls.Conn
	connOnce   sync.Once
	connMu     sync.Mutex
	serverAddr = "nenweb.supreme.com.tw:8080"
)

// InitConnection 初始化連接
func InitConnection() error {
	var err error
	connOnce.Do(func() {
		// 加載 CA 證書
		caCert, err := os.ReadFile("server.crt")
		if err != nil {
			return
		}

		// 創建 CA 證書池
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// 配置 TLS
		config := &tls.Config{
			RootCAs: caCertPool, // 設置 CA 證書池
			// InsecureSkipVerify: true, // 在生產環境中應設置為 false
		}

		// 建立 TLS 連接
		conn, err = tls.Dial("tcp", serverAddr, config)
	})

	return err
}

// CloseConnection 斷開連接
func CloseConnection() {
	if conn != nil {
		conn.Close()
		conn = nil
	}
}

// SendRequest 傳送請求
func SendRequest(req interface{}) (*json.Decoder, error) {
	if conn == nil {
		if err := InitConnection(); err != nil {
			return nil, fmt.Errorf("failed to initialize connection: %v", err)
		}
	}

	connMu.Lock()
	defer connMu.Unlock()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}

	return json.NewDecoder(conn), nil
}
