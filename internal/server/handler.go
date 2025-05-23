package server

import (
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
		response := models.Response{
			Error:     "Invalid station information",
			ErrorCode: models.ErrClientInvalidStation,
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
		return models.Response{
			Error:     fmt.Sprintf("Error marshaling request: %v", err),
			ErrorCode: models.ErrServerJSONMarshal,
		}
	}

	// 建立請求資料夾路徑
	reqDirPath := filepath.Join(basePath, "requests")
	if _, err := os.Stat(reqDirPath); os.IsNotExist(err) {
		err = os.Mkdir(reqDirPath, 0755)
		if err != nil {
			log.Printf("[ERROR %d] Error creating requests directory: %v", models.ErrServerCreateDir, err)
			return models.Response{
				Error:     fmt.Sprintf("Error creating requests directory: %v", err),
				ErrorCode: models.ErrServerCreateDir,
			}
		}
	}

	// 將 .json 檔案存放到指定請求資料夾
	reqFilePath := filepath.Join(reqDirPath, fmt.Sprintf("%s.json", reqUUID))
	err = os.WriteFile(reqFilePath, jsonData, 0644)
	if err != nil {
		log.Printf("[ERROR %d] Error writing request file: %v", models.ErrServerWriteFile, err)
		return models.Response{
			Error:     fmt.Sprintf("Error writing request file: %v", err),
			ErrorCode: models.ErrServerWriteFile,
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
				return models.Response{
					Error:     fmt.Sprintf("Error reading response file: %v", err),
					ErrorCode: models.ErrServerReadFile,
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
	return models.Response{
		Error:     "Timeout waiting for response file",
		ErrorCode: models.ErrServerTimeoutResponseFile,
	}
}

func ProcessMainAppRequest(req models.Request, basePath string) models.Response {
	// 建構獲取 .json 檔案資訊的路徑
	path := filepath.Join(basePath, req.AskId, req.AskContent.ProductSeries, req.AskContent.ApplicableProject)
	if req.AskContent.CustomizeId != "" {
		path = filepath.Join(path, "customize", req.AskContent.CustomizeId)
	}

	log.Printf("Searching for .json file in: %s", path)

	// 搜尋 .json 文件
	var jsonFile string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			jsonFile = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR %d] Error searching for json file: %v", models.ErrServerFileSearch, err)
		return models.Response{
			Error:     fmt.Sprintf("Error searching for json file: %v", err),
			ErrorCode: models.ErrServerFileSearch,
		}
	}

	if jsonFile == "" {
		log.Printf("[ERROR %d] No .json file found in: %s", models.ErrServerFileSearch, path)
		return models.Response{
			Error:     "No .json file found",
			ErrorCode: models.ErrServerFileSearch,
		}
	}

	// 檔案內容
	fileContent, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Printf("[ERROR %d] Error reading file content: %v", models.ErrServerReadFile, err)
		return models.Response{
			Error:     fmt.Sprintf("Error reading file content: %v", err),
			ErrorCode: models.ErrServerReadFile,
		}
	}

	var replyContent models.ReplyContent
	err = json.Unmarshal(fileContent, &replyContent)
	if err = json.Unmarshal(fileContent, &replyContent); err != nil {
		log.Printf("[ERROR %d] Error unmarshaling JSON data: %v", models.ErrServerJSONUnmarshal, err)
		return models.Response{
			Error:     fmt.Sprintf("Error unmarshaling JSON data: %v", err),
			ErrorCode: models.ErrServerJSONUnmarshal,
		}
	}

	// TODO: 這裡的回覆內容需要根據實際情況進行修改
	mainAppResponse := models.MainAppResponse{
		AskReplyId:      "MainApp",
		TimeStampClient: req.TimeStampClient,
		TimeStampServer: strconv.FormatInt(time.Now().Unix(), 10),
		ServerVersion:   "1.0.0",
		ReplyStatus:     "success",
		MsgValidateCode: "Msg",
		ReplyContent:    replyContent,
	}

	return models.Response{
		MainAppInfo: mainAppResponse,
	}
}
