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

// 檢查請求中站點資訊是否合法。資訊不完整回應 true, nil;完整回傳 false 及站點資訊
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

// 處理主程式請求
func ProcessMainAppRequest(req models.Request, basePath string) models.Response {
	return processRequestBase(req, basePath, ".bin", mainAppHandler)
}

// 處理參數表請求
func ProcessParameterTable(req models.Request, basePath string) models.Response {
	return processRequestBase(req, basePath, ".data", parameterTableHandler)
}

// 處理故障表請求
func ProcessFaultLogRequest(req models.Request, basePath string) models.Response {
	return models.Response{}
}

// 為請求建立檔案、監控回應並處理，適用不同的檔案類型
func processRequestBase(
	req models.Request,
	basePath string,
	fileExt string,
	handler models.FileResponseHandler,
) models.Response {
	// 唯一請求 ID
	reqUUID := uuid.New().String()
	// 設定 10 秒逾時 context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 建立來源路徑與請求內容
	sourcePath := buildSourcePath(req.AskContent)
	processReq := buildProcessRequest(req, sourcePath)
	processReqData, err := json.Marshal(processReq)
	if err != nil {
		log.Printf("[ERROR %d] Error marshaling request: %v", models.ErrServerJSONMarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONMarshal,
				ErrorMsg: fmt.Sprintf("Error marshaling request: %v", err),
			},
		}
	}

	// 準備請求與回應目錄
	reqDirPath := filepath.Join(basePath, "requests")
	respDirPath := filepath.Join(basePath, "responses")
	if err := ensureDirs(reqDirPath, respDirPath); err != nil {
		log.Printf("[ERROR %d] Error creating directory: %v", models.ErrServerCreateDir, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerCreateDir,
				ErrorMsg: fmt.Sprintf("Error creating directory: %v", err),
			},
		}
	}

	// 寫入請求檔案
	reqFilePath := filepath.Join(reqDirPath, fmt.Sprintf("%s.json", reqUUID))
	if err := writeRequestFile(reqFilePath, processReqData); err != nil {
		log.Printf("[ERROR %d] Error writing request file: %v", models.ErrServerWriteFile, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerWriteFile,
				ErrorMsg: fmt.Sprintf("Error writing request file: %v", err),
			},
		}
	}
	// 請求結束後自動清理臨時檔案
	defer cleanupFiles(reqFilePath)

	// 準備回應檔案路徑
	respJsonFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.json", reqUUID))
	respFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s%s", reqUUID, fileExt))

	// 建立回應通道並啟動監控協程
	resultCh := make(chan models.Response, 1)
	go monitorResponseFilesGeneric(ctx, respJsonFilePath, respFilePath, req, resultCh, handler)

	// 等待回應或逾時
	select {
	case result := <-resultCh:
		return result
	case <-ctx.Done():
		log.Printf("[ERROR %d] Timeout waiting for response files", models.ErrServerTimeoutResponseFile)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerTimeoutResponseFile,
				ErrorMsg: "Timeout waiting for response files",
			},
		}
	}
}

// 監控回應檔案，確保同時獲取 JSON 及資料檔案後，呼叫 handler 處理並回傳結果。
func monitorResponseFilesGeneric(
	ctx context.Context,
	jsonPath, dataPath string,
	req models.Request,
	resultCh chan<- models.Response,
	handler models.FileResponseHandler,
) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var (
		jsonData     []byte
		data         []byte
		jsonFileInfo os.FileInfo
		dataFileInfo os.FileInfo
		jsonErr      error
		dataErr      error
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 檢查 JSON 檔案是否存在且可讀取
			if jsonData == nil {
				jsonFileInfo, jsonErr = os.Stat(jsonPath)
				if jsonErr == nil && !jsonFileInfo.IsDir() {
					jsonData, jsonErr = os.ReadFile(jsonPath)
					if jsonErr != nil {
						log.Printf("[WARN] Error reading JSON file: %v, will retry", jsonErr)
						jsonData = nil
					}
				}
			}
			// 檢查資料檔案是否存在且可讀取
			if data == nil {
				dataFileInfo, dataErr = os.Stat(dataPath)
				if dataErr == nil && !dataFileInfo.IsDir() {
					data, dataErr = os.ReadFile(dataPath)
					if dataErr != nil {
						log.Printf("[WARN] Error reading DATA file: %v, will retry", dataErr)
						data = nil
					}
				}
			}
			//  兩檔案皆取得後，執行 handler 並回傳結果
			if jsonData != nil && data != nil {
				resultCh <- handler(jsonData, data, jsonFileInfo, dataFileInfo, req)
				return
			}
		}
	}
}

// 處理主應用程式回應檔案，將資料組裝成 MainAppResponse 結構。
func mainAppHandler(
	jsonData, binData []byte,
	jsonFileInfo, binFileInfo os.FileInfo,
	req models.Request,
) models.Response {
	var status models.Status
	// 解析 JSON 取得狀態
	if err := json.Unmarshal(jsonData, &status); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data: %v", err),
			},
		}
	}
	// 若狀態碼非 0，直接回傳錯誤
	if status.Code != 0 {
		log.Printf("[ERROR %d] Server returned error: %s", status.Code, status.ErrorMsg)
		return models.Response{Status: status}
	}
	var replyContent models.MainAppReplyContent
	// 解析 JSON 取得回應內容
	if err := json.Unmarshal(jsonData, &replyContent); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data into MainAppReplyContent: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data into MainAppReplyContent: %v", err),
			},
		}
	}
	// 補充檔案相關資訊
	replyContent.FileName = binFileInfo.Name()
	replyContent.FileSize = binFileInfo.Size()
	replyContent.FileType = ".bin"
	replyContent.AppContent = binData

	mainAppResponse := models.MainAppResponse{
		AskReplyId:      "MainApp",
		TimeStampClient: req.TimeStampClient,
		TimeStampServer: strconv.FormatInt(time.Now().Unix(), 10),
		ServerVersion:   "1.0.0",
		MsgValidateCode: "Msg",
		ReplyContent:    replyContent,
	}
	return models.Response{
		MainAppInfo: mainAppResponse,
		Status:      status,
	}
}

// 處理參數表回應檔案，將資料組裝成 ParameterTableResponse 結構。
func parameterTableHandler(
	jsonData, dataData []byte,
	jsonFileInfo, dataFileInfo os.FileInfo,
	req models.Request,
) models.Response {
	var status models.Status
	// 解析 JSON 取得狀態
	if err := json.Unmarshal(jsonData, &status); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data: %v", err),
			},
		}
	}
	// 若狀態碼非 0，直接回傳錯誤
	if status.Code != 0 {
		log.Printf("[ERROR %d] Server returned error: %s", status.Code, status.ErrorMsg)
		return models.Response{Status: status}
	}
	var replyContent models.ParameterTableReplyContent
	// 解析 JSON 取得回應內容
	if err := json.Unmarshal(jsonData, &replyContent); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data into ParameterTableReplyContent: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data into ParameterTableReplyContent: %v", err),
			},
		}
	}
	// 補充檔案相關資訊
	replyContent.FileName = dataFileInfo.Name()
	replyContent.FileSize = dataFileInfo.Size()
	replyContent.FileType = ".data"
	replyContent.ParameterContent = dataData

	parameterTableResponse := models.ParameterTableResponse{
		AskReplyId:      "ParameterTable",
		TimeStampClient: req.TimeStampClient,
		TimeStampServer: strconv.FormatInt(time.Now().Unix(), 10),
		ServerVersion:   "1.0.0",
		MsgValidateCode: "Msg",
		ReplyContent:    replyContent,
	}
	return models.Response{
		ParameterTableInfo: parameterTableResponse,
		Status:             status,
	}
}

// buildSourcePath 根據 AskContent 組合出來源路徑字串。
func buildSourcePath(askContent models.AskContent) string {
	if askContent.CustomizeId != "" {
		// 若 CustomizeId 存在，會加上 customize 子目錄。
		return fmt.Sprintf("\\%s\\%s\\customize\\%s", askContent.ProductSeries, askContent.ApplicableProject, askContent.CustomizeId)
	} else {
		return fmt.Sprintf("\\%s\\%s", askContent.ProductSeries, askContent.ApplicableProject)
	}
}

// 組合 ProcessRequest 結構，將請求資訊與來源路徑封裝。
func buildProcessRequest(req models.Request, sourcePath string) models.ProcessRequest {
	return models.ProcessRequest{
		Command:         req.Command,
		TimeStamp:       req.TimeStampClient,
		AskId:           req.AskId,
		AskContent:      req.AskContent,
		Station:         req.Station,
		MsgValidateCode: req.MsgValidateCode,
		SourcePath:      sourcePath,
	}
}

// 確保所有指定目錄皆存在，若不存在則建立。
func ensureDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// 將資料寫入指定檔案路徑
func writeRequestFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0644)
}

// 清理指定的臨時檔案，若檔案不存在則略過。
func cleanupFiles(files ...string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			log.Printf("[WARN] Failed to clean up file %s: %v", file, err)
		}
	}
}
