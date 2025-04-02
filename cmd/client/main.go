package main

/*
#include <stdlib.h>
#include <stdint.h>

// 定義結構體（與 Go 中的定義一致）
typedef struct {
	char* fileName;
	char* filePath;
	int64_t fileSize;
	char* fileContent;
	size_t fileContentSize;
	char* errorMsg;
	int errorCode;
} BinFileInfo;

typedef struct {
	int64_t fileSize;
	char* appContent;
	char* version;
	char* blVersion;
} ReplyContent;

typedef struct {
	char* askReplyId;
	char* timeStampClient;
	char* timeStampServer;
	char* serverVersion;
	char* replyStauts;
	char* msgValidateCode;
	ReplyContent replyContent;
	char* errorMsg;
	int errorCode;
} MainAppInfo;

*/
import "C"
import (
	"encoding/json"
	"strconv"
	"time"
	"unsafe"

	"github.com/socket-communication/internal/client"
	"github.com/socket-communication/internal/models"
)

//export InitConnection
func InitConnection(stationType, stationName, stationId, operatorId *C.char) *C.char {
	err := client.GetClient().InitConnection()
	if err != nil {
		response := models.Response{
			Error:     err.Error(),
			ErrorCode: models.ErrClientConnectionInitFailed, // 初始化連接失敗通常與 TLS 相關
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	request := models.Request{
		Command: "connect",
		Station: models.Station{
			StationType: C.GoString(stationType),
			StationName: C.GoString(stationName),
			StationId:   C.GoString(stationId),
			OperatorId:  C.GoString(operatorId),
		},
	}
	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		response := models.Response{
			Error:     err.Error(),
			ErrorCode: models.ErrClientRequestSendFailed,
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		response := models.Response{
			Error:     err.Error(),
			ErrorCode: models.ErrClientResponseDecodeFailed,
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	if response.Error != "" {
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	return C.CString("Initialize successfully")
}

//export CloseConnection
func CloseConnection() *C.char {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		response := models.Response{
			Error:     "Not connected",
			ErrorCode: models.ErrClientNotConnected,
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	request := models.Request{
		Command: "disconnect",
	}
	_, err := client.GetClient().SendRequest(request)
	if err != nil {
		response := models.Response{
			Error:     err.Error(),
			ErrorCode: models.ErrClientRequestSendFailed,
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	if err := client.GetClient().CloseConnection(); err != nil {
		response := models.Response{
			Error:     err.Error(),
			ErrorCode: models.ErrClientConnectionCloseFailed,
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	return C.CString("disconnect")
}

//export GetBinFileInfo
func GetBinFileInfo(askId, productSeries, applicableProject, customizeId *C.char) C.BinFileInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.BinFileInfo{
			errorMsg:  C.CString("Not connected"),
			errorCode: C.int(models.ErrClientNotConnected),
		}
	}

	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)

	// AskContent
	askContent := models.AskContent{
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
	}

	request := models.Request{
		TimeStampClient: timeUnix,
		AskId:           C.GoString(askId),
		AskContent:      askContent,
		IsGetBinFile:    true,
		MsgValidateCode: "validateCode",
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.BinFileInfo{
			errorMsg:  C.CString(err.Error()),
			errorCode: C.int(models.ErrClientRequestSendFailed),
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.BinFileInfo{
			errorMsg:  C.CString(err.Error()),
			errorCode: C.int(models.ErrClientResponseDecodeFailed),
		}
	}

	if response.Error != "" {
		return C.BinFileInfo{
			errorMsg:  C.CString(response.Error),
			errorCode: C.int(response.ErrorCode),
		}
	}

	return C.BinFileInfo{
		fileName:        C.CString(response.FileInfo.FileName),
		filePath:        C.CString(response.FileInfo.FilePath),
		fileSize:        C.int64_t(response.FileInfo.FileSize),
		fileContent:     C.CString(string(response.FileInfo.FileContent)),
		fileContentSize: C.size_t(len(response.FileInfo.FileContent)),
	}
}

//export FreeBinFileInfo
func FreeBinFileInfo(info C.BinFileInfo) {
	if info.fileName != nil {
		C.free(unsafe.Pointer(info.fileName))
	}
	if info.filePath != nil {
		C.free(unsafe.Pointer(info.filePath))
	}
	if info.fileContent != nil {
		C.free(unsafe.Pointer(info.fileContent))
	}
	if info.errorMsg != nil {
		C.free(unsafe.Pointer(info.errorMsg))
	}
}

//export GetMainAppInfo
func GetMainAppInfo(productSeries, applicableProject, customizeId *C.char) C.MainAppInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.MainAppInfo{
			errorMsg:  C.CString("Not connected"),
			errorCode: C.int(models.ErrClientNotConnected),
		}
	}

	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)

	// AskContent
	askContent := models.AskContent{
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
	}

	request := models.Request{
		TimeStampClient: timeUnix,
		AskId:           "MainApp",
		AskContent:      askContent,
		IsGetBinFile:    false,
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.MainAppInfo{
			errorMsg:  C.CString(err.Error()),
			errorCode: C.int(models.ErrClientRequestSendFailed),
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.MainAppInfo{
			errorMsg:  C.CString(err.Error()),
			errorCode: C.int(models.ErrClientResponseDecodeFailed),
		}
	}

	if response.Error != "" {
		return C.MainAppInfo{
			errorMsg:  C.CString(response.Error),
			errorCode: C.int(response.ErrorCode),
		}
	}

	replyContent := C.ReplyContent{
		fileSize:   C.int64_t(response.MainAppInfo.ReplyContent.FileSize),
		appContent: C.CString(response.MainAppInfo.ReplyContent.AppContent),
		version:    C.CString(response.MainAppInfo.ReplyContent.Version),
		blVersion:  C.CString(response.MainAppInfo.ReplyContent.BLVersion),
	}

	return C.MainAppInfo{
		askReplyId:      C.CString(response.MainAppInfo.AskReplyId),
		timeStampClient: C.CString(response.MainAppInfo.TimeStampClient),
		timeStampServer: C.CString(response.MainAppInfo.TimeStampServer),
		serverVersion:   C.CString(response.MainAppInfo.ServerVersion),
		replyStauts:     C.CString(response.MainAppInfo.ReplyStatus),
		msgValidateCode: C.CString(response.MainAppInfo.MsgValidateCode),
		replyContent:    replyContent,
	}
}

func main() {}
