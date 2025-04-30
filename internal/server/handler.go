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
	return processRequestBase(req, basePath, ".bin", mainAppHandler)
}

func ProcessParameterTable(req models.Request, basePath string) models.Response {
	return processRequestBase(req, basePath, ".data", parameterTableHandler)
}

func ProcessFaultLogRequest(req models.Request, basePath string) models.Response {
	return models.Response{}
}

func processRequestBase(
	req models.Request,
	basePath string,
	fileExt string,
	handler models.FileResponseHandler,
) models.Response {
	reqUUID := uuid.New().String()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
	defer cleanupFiles(reqFilePath)

	respJsonFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s.json", reqUUID))
	respFilePath := filepath.Join(respDirPath, fmt.Sprintf("%s%s", reqUUID, fileExt))

	resultCh := make(chan models.Response, 1)
	go monitorResponseFilesGeneric(ctx, respJsonFilePath, respFilePath, req, resultCh, handler)

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
			// 檢查JSON檔案
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
			// 檢查資料檔案
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
			// 兩檔案都到齊
			if jsonData != nil && data != nil {
				resultCh <- handler(jsonData, data, jsonFileInfo, dataFileInfo, req)
				return
			}
		}
	}
}

// 監控回應檔案 - 確保同時獲取JSON和BIN檔案
func mainAppHandler(
	jsonData, binData []byte,
	jsonFileInfo, binFileInfo os.FileInfo,
	req models.Request,
) models.Response {
	var status models.Status
	if err := json.Unmarshal(jsonData, &status); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data: %v", err),
			},
		}
	}
	if status.Code != 0 {
		log.Printf("[ERROR %d] Server returned error: %s", status.Code, status.ErrorMsg)
		return models.Response{Status: status}
	}
	var replyContent models.MainAppReplyContent
	if err := json.Unmarshal(jsonData, &replyContent); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data into MainAppReplyContent: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data into MainAppReplyContent: %v", err),
			},
		}
	}
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

// 監控回應檔案 - 確保同時獲取JSON和BIN檔案
func parameterTableHandler(
	jsonData, dataData []byte,
	jsonFileInfo, dataFileInfo os.FileInfo,
	req models.Request,
) models.Response {
	var status models.Status
	if err := json.Unmarshal(jsonData, &status); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data: %v", err),
			},
		}
	}
	if status.Code != 0 {
		log.Printf("[ERROR %d] Server returned error: %s", status.Code, status.ErrorMsg)
		return models.Response{Status: status}
	}
	var replyContent models.ParameterTableReplyContent
	if err := json.Unmarshal(jsonData, &replyContent); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data into ParameterTableReplyContent: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Status: models.Status{
				Code:     models.ErrServerJSONUnmarshal,
				ErrorMsg: fmt.Sprintf("Error unmarshaling JSON data into ParameterTableReplyContent: %v", err),
			},
		}
	}
	replyContent.FileName = dataFileInfo.Name()
	replyContent.FileSize = dataFileInfo.Size()
	replyContent.FileType = ".data"
	replyContent.AppContent = dataData

	parameterTableResponse := models.ParameterTableResponse{
		AskReplyId:      "MainApp",
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

func buildSourcePath(askContent models.AskContent) string {
	if askContent.CustomizeId != "" {
		return fmt.Sprintf("\\%s\\%s\\customize\\%s", askContent.ProductSeries, askContent.ApplicableProject, askContent.CustomizeId)
	} else {
		return fmt.Sprintf("\\%s\\%s", askContent.ProductSeries, askContent.ApplicableProject)
	}
}

func buildProcessRequest(req models.Request, sourcePath string) models.ProcessRequest {
	return models.ProcessRequest{
		TimeStamp:       req.TimeStampClient,
		AskId:           req.AskId,
		AskContent:      req.AskContent,
		Station:         req.Station,
		MsgValidateCode: req.MsgValidateCode,
		SourcePath:      sourcePath,
	}
}

func ensureDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func writeRequestFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0644)
}

// 清理臨時檔案
func cleanupFiles(files ...string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			log.Printf("[WARN] Failed to clean up file %s: %v", file, err)
		}
	}
}
