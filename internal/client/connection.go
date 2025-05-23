package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type TLSClient struct {
	conn    *tls.Conn
	connMux sync.Mutex
}

var (
	client     *TLSClient
	clientOnce sync.Once
	serverAddr = "nenweb.supreme.com.tw:443"
)

func GetClient() *TLSClient {
	clientOnce.Do(func() {
		client = &TLSClient{}
	})
	return client
}

// InitConnection 初始化連接
func (c *TLSClient) InitConnection() error {
	c.connMux.Lock()
	defer c.connMux.Unlock()

	// 關閉現有連接（如果存在）
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	// 加載 CA 證書
	caCert, err := os.ReadFile("server.crt")
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}

	// 創建 CA 證書池，接受自簽憑證
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to parse CA certificate")
	}

	// 配置 TLS
	config := &tls.Config{
		RootCAs: caCertPool, // 設置 CA 證書池
	}

	// 建立 TLS 連接
	c.conn, err = tls.Dial("tcp", serverAddr, config)
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection: %v", err)
	}

	return nil
}

// CloseConnection 斷開連接
func (c *TLSClient) CloseConnection() error {
	c.connMux.Lock()
	defer c.connMux.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// SendRequest 傳送請求
func (c *TLSClient) SendRequest(req interface{}) (*json.Decoder, error) {
	c.connMux.Lock()
	defer c.connMux.Unlock()

	// 檢查連接狀態
	if c.conn == nil {
		if err := c.InitConnection(); err != nil {
			return nil, fmt.Errorf("failed to initialize connection: %v", err)
		}
	}

	// 發送請求
	encoder := json.NewEncoder(c.conn)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}

	// 返回解碼器
	return json.NewDecoder(c.conn), nil
}

// IsConnected 檢查是否已連接
func (c *TLSClient) IsConnected() bool {
	c.connMux.Lock()
	defer c.connMux.Unlock()
	return c.conn != nil
}
