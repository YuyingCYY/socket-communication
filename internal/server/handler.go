package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/socket-communication/internal/models"
)

func ValidateStationInfo(req *models.Request, encoder *json.Encoder) (bool, *models.Station) {
	if req.Station.StationName == "" || req.Station.StationType == "" || req.Station.StationId == "" || req.Station.OperatorId == "" {

		response := models.Status{
			Code:     models.ErrClientInvalidStation,
			ErrorMsg: "Invalid station information",
		}
		if err := encoder.Encode(response); err != nil {
			log.Printf("[ERROR %d]Error encoding response: %v", models.ErrClientInvalidStation, err)
		}
		return true, nil
	}

	return false, &req.Station
}

func ProcessFileRequest(req models.Request, basePath string) models.Response {
	// UUID 作為檔案名稱
	reqUUID := uuid.New().String()

	// 將請求內容轉換成json
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ERROR %d] Error marshaling request: %v", models.ErrServerJSONMarshal, err)
		status := models.Status{
			Code:     models.ErrServerJSONMarshal,
			ErrorMsg: fmt.Sprintf("Error marshaling request: %v", err),
		}
		return models.Response{
			Status: status,
		}
	}

	// 建立請求資料夾路徑
	reqDirPath := filepath.Join(basePath, "requests")
	if _, err := os.Stat(reqDirPath); os.IsNotExist(err) {
		err = os.Mkdir(reqDirPath, 0755)
		if err != nil {
			log.Printf("[ERROR %d] Error creating requests directory: %v", models.ErrServerCreateDir, err)
			status := models.Status{
				Code:     models.ErrServerCreateDir,
				ErrorMsg: fmt.Sprintf("Error creating requests directory: %v", err),
			}
			return models.Response{
				Status: status,
			}
		}
	}

	// 將 .json 檔案存放到指定請求資料夾
	reqFilePath := filepath.Join(reqDirPath, fmt.Sprintf("%s.json", reqUUID))
	err = os.WriteFile(reqFilePath, jsonData, 0644)
	if err != nil {
		log.Printf("[ERROR %d] Error writing request file: %v", models.ErrServerWriteFile, err)
		status := models.Status{
			Code:     models.ErrServerWriteFile,
			ErrorMsg: fmt.Sprintf("Error writing request file: %v", err),
		}
		return models.Response{
			Status: status,
		}
	}

	// 監聽並等待指定資料夾有回傳檔案
	respDirPath := filepath.Join(basePath, "responses")
	respFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.bin", reqUUID))
	for i := range 10 { // 最多等待 10 秒
		if fileInfo, err := os.Stat(respFilePath); !os.IsNotExist(err) {
			// 讀取回傳檔案內容
			respData, err := os.ReadFile(respFilePath)
			if err != nil {
				log.Printf("[ERROR %d] Error reading response file: %v", models.ErrServerReadFile, err)
				status := models.Status{
					Code:     models.ErrServerReadFile,
					ErrorMsg: fmt.Sprintf("Error reading response file: %v", err),
				}
				return models.Response{
					Status: status,
				}
			}

			// 將回傳檔案內容解析為 Response struct
			// var resp models.FileResponse
			// err = json.Unmarshal(respData, &resp)
			// if err != nil {
			// 	errorMsg = "Error unmarshaling response: " + err.Error()
			// 	log.Print(errorMsg)
			// 	return models.Response{
			// 		Error: errorMsg,
			// 	}
			// }

			fileResponse := models.FileResponse{
				FileName:    fileInfo.Name(),
				FilePath:    respFilePath,
				FileSize:    fileInfo.Size(),
				FileContent: respData,
			}

			return models.Response{
				FileInfo: fileResponse,
			}
		}

		i++
		time.Sleep(1 * time.Second)
	}

	log.Printf("[ERROR %d] Timeout waiting for response file", models.ErrServerTimeoutResponseFile)
	status := models.Status{
		Code:     models.ErrServerTimeoutResponseFile,
		ErrorMsg: "Timeout waiting for response file",
	}
	return models.Response{
		Status: status,
	}
}

// func ProcessMainAppRequest(req models.Request, basePath string) models.Response {
// 	// UUID 作為檔案名稱
// 	reqUUID := uuid.New().String()

// 	// 將請求內容轉換成json
// 	jsonData, err := json.Marshal(req)
// 	if err != nil {
// 		log.Printf("[ERROR %d] Error marshaling request: %v", models.ErrServerJSONMarshal, err)
// 		return models.Response{
// 			Error:     fmt.Sprintf("Error marshaling request: %v", err),
// 			ErrorCode: models.ErrServerJSONMarshal,
// 		}
// 	}

// 	// 建立請求資料夾路徑
// 	reqDirPath := filepath.Join(basePath, "requests")
// 	if _, err := os.Stat(reqDirPath); os.IsNotExist(err) {
// 		err = os.Mkdir(reqDirPath, 0755)
// 		if err != nil {
// 			log.Printf("[ERROR %d] Error creating requests directory: %v", models.ErrServerCreateDir, err)
// 			return models.Response{
// 				Error:     fmt.Sprintf("Error creating requests directory: %v", err),
// 				ErrorCode: models.ErrServerCreateDir,
// 			}
// 		}
// 	}

// 	// 將 .json 檔案存放到指定請求資料夾
// 	reqFilePath := filepath.Join(reqDirPath, fmt.Sprintf("%s.json", reqUUID))
// 	err = os.WriteFile(reqFilePath, jsonData, 0644)
// 	if err != nil {
// 		log.Printf("[ERROR %d] Error writing request file: %v", models.ErrServerWriteFile, err)
// 		return models.Response{
// 			Error:     fmt.Sprintf("Error writing request file: %v", err),
// 			ErrorCode: models.ErrServerWriteFile,
// 		}
// 	}

// 	// 監聽並等待指定資料夾有回傳檔案
// 	respDirPath := filepath.Join(basePath, "responses")
// 	respBinFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.bin", reqUUID))
// 	respJsonFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.json", reqUUID))
// 	var replyContent models.MainAppReplyContent
// 	for i := range 10 { // 最多等待 10 秒
// 		if _, err := os.Stat(respJsonFilePath); !os.IsNotExist(err) {
// 			fileContent, err := os.ReadFile(respJsonFilePath)
// 			if err != nil {
// 				log.Printf("[ERROR %d] Error reading response JSON file: %v", models.ErrServerReadFile, err)
// 				return models.Response{
// 					Error:     fmt.Sprintf("Error reading response JSON file: %v", err),
// 					ErrorCode: models.ErrServerReadFile,
// 				}
// 			}

// 			if err = json.Unmarshal(fileContent, &replyContent); err != nil {
// 				log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
// 				return models.Response{
// 					Error:     fmt.Sprintf("Error unmarshaling JSON data: %v", err),
// 					ErrorCode: models.ErrServerJSONUnmarshal,
// 				}
// 			}
// 		}

// 		if binFileInfo, err := os.Stat(respBinFilePath); !os.IsNotExist(err) {
// 			// 讀取回傳檔案內容
// 			respData, err := os.ReadFile(respBinFilePath)
// 			if err != nil {
// 				log.Printf("[ERROR %d] Error reading response file: %v", models.ErrServerReadFile, err)
// 				return models.Response{
// 					Error:     fmt.Sprintf("Error reading response file: %v", err),
// 					ErrorCode: models.ErrServerReadFile,
// 				}
// 			}

// 			replyContent.FileName = binFileInfo.Name()
// 			replyContent.FileSize = binFileInfo.Size()
// 			replyContent.FileType = ".bin"
// 			replyContent.FileContent = respData

// 			mainAppResponse := models.MainAppResponse{
// 				AskReplyId:      "MainApp",
// 				TimeStampClient: req.TimeStampClient,
// 				TimeStampServer: strconv.FormatInt(time.Now().Unix(), 10),
// 				ServerVersion:   "1.0.0",
// 				ReplyStatus:     "success",
// 				MsgValidateCode: "Msg",
// 				ReplyContent:    replyContent,
// 			}

// 			return models.Response{
// 				MainAppInfo: mainAppResponse,
// 			}
// 		}

// 		i++
// 		time.Sleep(1 * time.Second)
// 	}

// 	log.Printf("[ERROR %d] Timeout waiting for response file", models.ErrServerTimeoutResponseFile)
// 	return models.Response{
// 		Error:     "Timeout waiting for response file",
// 		ErrorCode: models.ErrServerTimeoutResponseFile,
// 	}
// }

func ProcessMainAppRequest(req models.Request, basePath string) models.Response {
	// UUID 作為檔案名稱
	reqUUID := uuid.New().String()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 將請求內容轉換成json
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ERROR %d] Error marshaling request: %v", models.ErrServerJSONMarshal, err)
		status := models.Status{
			Code:     models.ErrServerJSONMarshal,
			ErrorMsg: fmt.Sprintf("Error marshaling request: %v", err),
		}
		return models.Response{
			Status: status,
		}
	}

	// 確保請求和回應資料夾存在
	reqDirPath := filepath.Join(basePath, "requests")
	respDirPath := filepath.Join(basePath, "responses")
	for _, dirPath := range []string{reqDirPath, respDirPath} {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Printf("[ERROR %d] Error creating directory %s: %v", models.ErrServerCreateDir, dirPath, err)
			status := models.Status{
				Code:     models.ErrServerCreateDir,
				ErrorMsg: fmt.Sprintf("Error creating directory %s: %v", dirPath, err),
			}
			return models.Response{
				Status: status,
			}
		}
	}

	// 將 .json 檔案存放到指定請求資料夾
	reqFilePath := filepath.Join(reqDirPath, fmt.Sprintf("%s.json", reqUUID))
	err = os.WriteFile(reqFilePath, jsonData, 0644)
	if err != nil {
		log.Printf("[ERROR %d] Error writing request file: %v", models.ErrServerWriteFile, err)
		status := models.Status{
			Code:     models.ErrServerWriteFile,
			ErrorMsg: fmt.Sprintf("Error writing request file: %v", err),
		}
		return models.Response{
			Status: status,
		}
	}
	defer cleanupFiles(reqFilePath)

	// 監聽並等待指定資料夾有回傳檔案
	respJsonFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.json", reqUUID))
	respBinFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.bin", reqUUID))

	// 使用channel來接收回應結果
	resultCh := make(chan models.Response, 1)

	// 啟動檔案監控
	go monitorResponseFiles(ctx, respJsonFilePath, respBinFilePath, req, resultCh)

	// 等待回應或超時
	select {
	case result := <-resultCh:
		// 清理回應檔案
		// cleanupFiles(respJsonFilePath, respBinFilePath)
		return result
	case <-ctx.Done():
		log.Printf("[ERROR %d] Timeout waiting for response files", models.ErrServerTimeoutResponseFile)
		status := models.Status{
			Code:     models.ErrServerTimeoutResponseFile,
			ErrorMsg: "Timeout waiting for response files",
		}
		return models.Response{
			Status: status,
		}
	}
}

// 監控回應檔案 - 確保同時獲取JSON和BIN檔案
func monitorResponseFiles(ctx context.Context, jsonPath, binPath string, req models.Request, resultCh chan<- models.Response) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var (
		jsonData     []byte
		binData      []byte
		jsonFileInfo os.FileInfo
		binFileInfo  os.FileInfo
		jsonErr      error
		binErr       error
	)

	for {
		select {
		case <-ctx.Done():
			// 上下文已取消（超時）
			return
		case <-ticker.C:
			// 檢查JSON檔案存在
			if jsonData == nil {
				jsonFileInfo, jsonErr = os.Stat(jsonPath)
				if jsonErr == nil && !jsonFileInfo.IsDir() {
					jsonData, jsonErr = os.ReadFile(jsonPath)
					if jsonErr != nil {
						log.Printf("[WARN] Error reading JSON file: %v, will retry", jsonErr)
						jsonData = nil // 重置以便重試
					} else {
						log.Printf("JSON file found: %s", jsonPath)
					}
				}
			}

			// 檢查BIN檔案存在
			if binData == nil {
				binFileInfo, binErr = os.Stat(binPath)
				if binErr == nil && !binFileInfo.IsDir() {
					binData, binErr = os.ReadFile(binPath)
					if binErr != nil {
						log.Printf("[WARN] Error reading BIN file: %v, will retry", binErr)
						binData = nil // 重置以便重試
					} else {
						log.Printf("BIN file found: %s", binPath)
					}
				}
			}

			// 只有當兩個檔案都成功讀取才處理回應
			if jsonData != nil && binData != nil {
				var replyContent models.MainAppReplyContent

				// 解析JSON內容
				if err := json.Unmarshal(jsonData, &replyContent); err != nil {
					log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
					status := models.Status{
						Code:     models.ErrServerJSONUnmarshal,
						ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data: %v", err),
					}

					resultCh <- models.Response{
						Status: status,
					}
					return
				}

				// 設定BIN檔案信息
				replyContent.FileName = binFileInfo.Name()
				replyContent.FileSize = binFileInfo.Size()
				replyContent.FileType = ".bin"
				replyContent.AppContent = binData

				// 構建完整回應
				mainAppResponse := models.MainAppResponse{
					AskReplyId:      "MainApp",
					TimeStampClient: req.TimeStampClient,
					TimeStampServer: strconv.FormatInt(time.Now().Unix(), 10),
					ServerVersion:   "1.0.0",
					ReplyStatus:     "success",
					MsgValidateCode: "Msg",
					ReplyContent:    replyContent,
				}

				resultCh <- models.Response{
					MainAppInfo: mainAppResponse,
				}
				return
			} else {
				// 記錄當前狀態 (可選，可能會產生大量日誌)
				// if ctx.Err() == nil { // 只有在未超時的情況下才記錄
				// 	log.Printf("Waiting for files - JSON: %v, BIN: %v",
				// 		jsonData != nil, binData != nil)
				// }
			}
		}
	}
}

// 清理臨時檔案
func cleanupFiles(files ...string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			log.Printf("[WARN] Failed to clean up file %s: %v", file, err)
		}
	}
}
