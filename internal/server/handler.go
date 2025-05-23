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

	// 同時搜尋.json和.bin檔案
	var jsonFile, binFile string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".json" {
			jsonFile = path
		} else if ext == ".bin" {
			binFile = path
		}
		return nil
	})

	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error searching for json file: %v", err),
		}
	}

	// 檢查必要檔案存在
	if binFile == "" {
		return models.Response{Error: "No .bin file found."}
	}
	if jsonFile == "" {
		return models.Response{
			Error: "No .json file found",
		}
	}

	// 讀取JSON檔案
	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		return models.Response{Error: fmt.Sprintf("JSON檔案讀取失敗: %v", err)}
	}

	var replyContent models.MainAppReplyContent
	if err := json.Unmarshal(jsonContent, &replyContent); err != nil {
		return models.Response{Error: fmt.Sprintf("JSON解析錯誤: %v", err)}
	}

	// 讀取BIN檔案並更新內容
	binData, err := os.ReadFile(binFile)
	if err != nil {
		return models.Response{Error: fmt.Sprintf("BIN檔案讀取失敗: %v", err)}
	}

	// 取得檔案資訊
	binInfo, err := os.Stat(binFile)
	if err != nil {
		return models.Response{Error: fmt.Sprintf("BIN檔案資訊取得失敗: %v", err)}
	}

	// 更新回覆內容
	replyContent.FileSize = binInfo.Size()
	replyContent.AppContent = binData

	// 建構回應
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

func ProcessParameterTableRequest(req models.Request, basePath string) models.Response {
	// 建構獲取 .json 檔案資訊的路徑
	path := filepath.Join(basePath, req.AskId, req.AskContent.ProductSeries, req.AskContent.ApplicableProject)
	if req.AskContent.CustomizeId != "" {
		path = filepath.Join(path, "customize", req.AskContent.CustomizeId)
	}

	log.Printf("Searching for .json file in: %s", path)

	// 同時搜尋.json和.bin檔案
	var jsonFile, dataFile string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".json" {
			jsonFile = path
		} else if ext == ".data" {
			dataFile = path
		}
		return nil
	})

	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error searching for json file: %v", err),
		}
	}

	// 檢查必要檔案存在
	if dataFile == "" {
		return models.Response{Error: "No .bin file found."}
	}
	if jsonFile == "" {
		return models.Response{
			Error: "No .json file found",
		}
	}

	// 讀取JSON檔案
	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		return models.Response{Error: fmt.Sprintf("JSON檔案讀取失敗: %v", err)}
	}

	var replyContent models.ParameterTableReplyContent
	if err := json.Unmarshal(jsonContent, &replyContent); err != nil {
		return models.Response{Error: fmt.Sprintf("JSON解析錯誤: %v", err)}
	}

	// 讀取BIN檔案並更新內容
	dataData, err := os.ReadFile(dataFile)
	if err != nil {
		return models.Response{Error: fmt.Sprintf("BIN檔案讀取失敗: %v", err)}
	}

	// 取得檔案資訊
	binInfo, err := os.Stat(dataFile)
	if err != nil {
		return models.Response{Error: fmt.Sprintf("BIN檔案資訊取得失敗: %v", err)}
	}

	// 更新回覆內容
	replyContent.FileSize = binInfo.Size()
	replyContent.ParameterContent = dataData

	// 建構回應
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
	}
}
