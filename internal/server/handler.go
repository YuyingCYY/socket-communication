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

func ProcessMainAppRequest(req models.Request, basePath string) models.Response {
	// UUID 作為檔案名稱
	reqUUID := uuid.New().String()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 彙整並建立要存放到請求資料夾的Data
	askContent := req.AskContent
	var sourcePath string
	if askContent.CustomizeId != "" {
		sourcePath = fmt.Sprintf("\\%s\\%s\\customize\\%s", askContent.ProductSeries, askContent.ApplicableProject, askContent.CustomizeId)
	} else {
		sourcePath = fmt.Sprintf("\\%s\\%s", askContent.ProductSeries, askContent.ApplicableProject)
	}
	processReq := models.ProcessRequest{
		TimeStamp:       req.TimeStampClient,
		AskId:           req.AskId,
		AskContent:      req.AskContent,
		Station:         req.Station,
		MsgValidateCode: req.MsgValidateCode,
		SourcePath:      sourcePath,
	}

	// 將請求內容轉換成json
	processReqData, err := json.Marshal(processReq)
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
	err = os.WriteFile(reqFilePath, processReqData, 0644)
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
				// 先檢查 JSON 是否包含錯誤代碼
				var status models.Status

				// 解析JSON內容
				if err := json.Unmarshal(jsonData, &status); err != nil {
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

				// 檢查代碼是否為非零(錯誤)
				if status.Code != 0 {
					log.Printf("[ERROR %d] Server returned error: %s", status.Code, status.ErrorMsg)
					resultCh <- models.Response{
						Status: status,
					}
					return
				}

				// 若為 0 表示成功
				var replyContent models.MainAppReplyContent
				if err := json.Unmarshal(jsonData, &replyContent); err != nil {
					log.Printf("[ERROR %d] Error unmarshaling JSON data into MainAppReplyContent: %v", models.ErrServerJSONUnmarshal, err)
					status := models.Status{
						Code:     models.ErrServerJSONUnmarshal,
						ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data into MainAppReplyContent: %v", err),
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
					MsgValidateCode: "Msg",
					ReplyContent:    replyContent,
				}

				resultCh <- models.Response{
					MainAppInfo: mainAppResponse,
					Status:      status,
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
