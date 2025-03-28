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
	log.Printf("Station name: %s, Station Type: %s, Station ID: %s, Operator ID: %s",
		req.Station.StationName, req.Station.StationType, req.Station.StationId, req.Station.OperatorId)

	if req.Station.StationName == "" || req.Station.StationType == "" || req.Station.StationId == "" || req.Station.OperatorId == "" {
		response := models.Response{
			Error: "Invalid station information",
		}
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return true, nil
	}

	return false, &req.Station
}

func ProcessFileRequest(req models.Request, basePath string) models.Response {
	// UUID 作為檔案名稱
	reqUUID := uuid.New().String()
	var errorMsg string

	// 將請求內容轉換成json
	jsonData, err := json.Marshal(req)
	if err != nil {
		errorMsg = "Error marshaling request: " + err.Error()
		log.Print(errorMsg)
		return models.Response{
			Error: errorMsg,
		}
	}

	// 建立請求資料夾路徑
	reqDirPath := filepath.Join(basePath, "requests")
	if _, err := os.Stat(reqDirPath); os.IsNotExist(err) {
		err = os.Mkdir(reqDirPath, 0755)
		if err != nil {
			errorMsg = "Error creating requests directory: " + err.Error()
			log.Print(errorMsg)
			return models.Response{
				Error: errorMsg,
			}
		}
	}

	// 將 .json 檔案存放到指定請求資料夾
	reqFilePath := filepath.Join(reqDirPath, fmt.Sprintf("%s.json", reqUUID))
	err = os.WriteFile(reqFilePath, jsonData, 0644)
	if err != nil {
		errorMsg = "Error writing request file: " + err.Error()
		log.Print(errorMsg)
		return models.Response{
			Error: errorMsg,
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
				errorMsg = "Error reading response file: " + err.Error()
				log.Print(errorMsg)
				return models.Response{
					Error: errorMsg,
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

	// 回傳 Response
	errorMsg = "Timeout waiting for response file"
	log.Print(errorMsg)
	return models.Response{
		Error: errorMsg,
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
		return models.Response{
			Error: fmt.Sprintf("Error searching for json file: %v", err),
		}
	}

	if jsonFile == "" {
		return models.Response{
			Error: "No .json file found",
		}
	}

	// 檔案內容
	fileContent, err := os.ReadFile(jsonFile)
	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error reading file contetn: %v", err),
		}
	}

	var replyContent models.ReplyContent
	err = json.Unmarshal(fileContent, &replyContent)

	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error Unmarshal parses the JSON data: %v", err),
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
