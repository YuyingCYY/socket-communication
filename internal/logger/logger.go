package logger

import (
	"log"
	"os"
	"path/filepath"
)

func InitLogger(logFile string) error {
	// 確保日誌目錄存在
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 打開日誌文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return nil
}
