package logger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile   *os.File
	lastDate  string
	logDir    string
	initMutex sync.Mutex
)

// 初始化日誌系統
func InitLogger(dir string) error {
	initMutex.Lock()
	defer initMutex.Unlock()

	logDir = dir

	// 確保日誌目錄存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	return updateLogFile()
}

// updateLogFile 更新日誌文件
func updateLogFile() error {
	// 根據當前日期生成日誌文件名
	currentDate := time.Now().Format("2006-01-02")
	if currentDate == lastDate && logFile != nil {
		return nil // 日期未變化，無需更新
	}

	// 關閉舊的日誌文件
	if logFile != nil {
		logFile.Close()
	}

	// 創建新的日誌文件
	logFileName := filepath.Join(logDir, "log-"+currentDate+".log")
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// 設置日誌輸出
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 更新狀態
	logFile = file
	lastDate = currentDate

	return nil
}

// 啟動定時切換日誌檔案的 goroutine
func StartLogFileRotator() {
	go func() {
		for {
			time.Sleep(time.Minute)
			initMutex.Lock()
			_ = updateLogFile() // 忽略錯誤處理
			initMutex.Unlock()
		}
	}()
}

// Log 寫入日誌，自動檢查是否需要切換日誌文件
func Log(message string) {
	initMutex.Lock()
	defer initMutex.Unlock()

	// 檢查是否需要切換日誌文件
	if err := updateLogFile(); err != nil {
		log.Println("Failed to update log file:", err)
	}

	// 寫入日誌
	log.Println(message)
}
