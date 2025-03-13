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

//export InitializeClient
func InitializeClient(stationType, stationName, stationId, operatorId *C.char) bool {
	err := client.GetClient().InitConnection()
	if err != nil {
		return false
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
			Error: err.Error(),
		}
		_, _ = json.Marshal(response)
		return false
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return false
	}

	if response.Error != "" {
		return false
	}

	return true
}

//export CloseConnection
func CloseConnection() bool {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		response := models.Response{
			Error: "Not connected",
		}
		_, _ = json.Marshal(response)
		return false
	}

	request := models.Request{
		Command: "disconnect",
	}
	_, err := client.GetClient().SendRequest(request)
	if err != nil {
		response := models.Response{
			Error: err.Error(),
		}
		_, _ = json.Marshal(response)
		return false
	}

	if err := client.GetClient().CloseConnection(); err != nil {
		response := models.Response{
			Error: err.Error(),
		}
		_, _ = json.Marshal(response)
		return false
	}

	return true
}

//export GetBinFileInfo
func GetBinFileInfo(askId, productSeries, applicableProject, customizeId *C.char) C.BinFileInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.BinFileInfo{
			errorMsg: C.CString("Not connected"),
		}
	}

	request := models.Request{
		AskId:             C.GoString(askId),
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
		IsGetBinFile:      true,
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.BinFileInfo{
			errorMsg: C.CString(err.Error()),
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.BinFileInfo{
			errorMsg: C.CString(err.Error()),
		}
	}

	if response.Error != "" {
		return C.BinFileInfo{
			errorMsg: C.CString(response.Error),
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
			errorMsg: C.CString("Not connected"),
		}
	}

	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)
	request := models.Request{
		TimeStampClient:   timeUnix,
		AskId:             "MainApp",
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
		IsGetBinFile:      false,
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.MainAppInfo{
			errorMsg: C.CString(err.Error()),
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.MainAppInfo{
			errorMsg: C.CString(err.Error()),
		}
	}

	if response.Error != "" {
		return C.MainAppInfo{
			errorMsg: C.CString(response.Error),
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
