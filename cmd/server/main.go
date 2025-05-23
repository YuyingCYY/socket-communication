package main

import (
	"log"

	"github.com/socket-communication/internal/config"
	"github.com/socket-communication/internal/logger"
)

func main() {
	cfg := config.LoadConfig()

	if err := logger.InitLogger(cfg.LogFile); err != nil {
		panic(err)
	}
	logger.StartLogFileRotator()

	// 建立並啟動伺服器
	srv := NewSocketServer(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
