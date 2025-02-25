package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/socket-communication/internal/models"
)

func ProcessFileRequest(req models.Request, basePath string) models.Response {
	// 建構完整路徑
	path := filepath.Join(basePath, req.AskId, req.ProductSeries, req.ApplicableProject)
	if req.CustomizeId != "" {
		path = filepath.Join(path, "customize", req.CustomizeId)
	}

	log.Printf("Searching for .bin file in: %s", path)

	// 搜尋 .bin 文件
	var binFile string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".bin" {
			binFile = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error searching for bin file: %v", err),
		}
	}

	if binFile == "" {
		return models.Response{
			Error: "No .bin file found",
		}
	}

	// 檔案訊息
	fileInfo, err := os.Stat(binFile)
	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error getting file info: %v", err),
		}
	}

	// 檔案內容
	fileContent, err := os.ReadFile(binFile)
	if err != nil {
		return models.Response{
			Error: fmt.Sprintf("Error reading file contetn: %v", err),
		}
	}

	fileResponse := models.FileResponse{
		FileName:    fileInfo.Name(),
		FilePath:    binFile,
		FileSize:    fileInfo.Size(),
		FileContent: fileContent,
	}

	return models.Response{
		FileInfo: fileResponse,
	}
}

func ProcessMainAppRequest(req models.Request, basePath string) models.Response {
	// 建構獲取 .json 檔案資訊的路徑
	path := filepath.Join(basePath, req.AskId, req.ProductSeries, req.ApplicableProject)
	if req.CustomizeId != "" {
		path = filepath.Join(path, "customize", req.CustomizeId)
	}

	log.Printf("Searching for .bin file in: %s", path)

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
			Error: fmt.Sprintf("Error Unmarshal parses the JSON data"),
		}
	}

	mainAppResponse := models.MainAppResponse{
		AskReplyId:      "MainApp",
		TimeStampClient: "client",
		TimeStampServer: "server",
		ServerVersion:   "1.0.0",
		ReplyStatus:     "success",
		MsgValidateCode: "Msg",
		ReplyContent:    replyContent,
	}

	return models.Response{
		MainAppInfo: mainAppResponse,
	}
}
