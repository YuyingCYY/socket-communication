package main

/*
#include <stdlib.h>
#include <stdint.h>

// 定義結構體（與 Go 中的定義一致）
typedef struct {
	int code;
	char* errorMsg;
} Status;

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
	char* fileName;
	int64_t fileSize;
	char* fileType;
	char* appContent;
	char* version;
	char* blVersion;
} MainAppReplyContent;

typedef struct {
	char* askReplyId;
	char* timeStampClient;
	char* timeStampServer;
	char* serverVersion;
	char* msgValidateCode;
	MainAppReplyContent replyContent;
	Status status;
} MainAppInfo;

*/
import "C"
import (
	"strconv"
	"time"
	"unsafe"

	"github.com/socket-communication/internal/client"
	"github.com/socket-communication/internal/models"
)

//export InitConnection
func InitConnection(stationType, stationName, stationId, operatorId *C.char) C.Status {
	err := client.GetClient().InitConnection()
	if err != nil {
		return C.Status{
			code:     models.ErrClientConnectionInitFailed, // 初始化連接失敗通常與 TLS 相關
			errorMsg: C.CString(err.Error()),
		}
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
		return C.Status{
			code:     models.ErrClientRequestSendFailed,
			errorMsg: C.CString(err.Error()),
		}
	}

	var response models.Status
	if err := decoder.Decode(&response); err != nil {
		return C.Status{
			code:     models.ErrClientResponseDecodeFailed,
			errorMsg: C.CString(err.Error()),
		}
	}

	return C.Status{
		code: 0,
	}
}

//export CloseConnection
func CloseConnection() C.Status {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.Status{
			code:     models.ErrClientNotConnected,
			errorMsg: C.CString("Not connected"),
		}
	}

	request := models.Request{
		Command: "disconnect",
	}
	_, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.Status{
			code:     models.ErrClientRequestSendFailed,
			errorMsg: C.CString(err.Error()),
		}
	}

	if err := client.GetClient().CloseConnection(); err != nil {
		return C.Status{
			code:     models.ErrClientConnectionCloseFailed,
			errorMsg: C.CString(err.Error()),
		}
	}

	return C.Status{
		code: 0,
	}
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

	if response.Status.ErrorMsg != "" {
		return C.BinFileInfo{
			errorMsg:  C.CString(response.Status.ErrorMsg),
			errorCode: C.int(response.Status.Code),
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
			status: C.Status{
				code:     C.int(models.ErrClientNotConnected),
				errorMsg: C.CString("Not connected"),
			},
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
			status: C.Status{
				code:     C.int(models.ErrClientRequestSendFailed),
				errorMsg: C.CString(err.Error()),
			},
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.MainAppInfo{
			status: C.Status{
				code:     C.int(models.ErrClientResponseDecodeFailed),
				errorMsg: C.CString(err.Error()),
			},
		}
	}

	// 代碼不為`0`皆為錯誤碼
	if response.Status.Code != 0 {
		return C.MainAppInfo{
			status: C.Status{
				code:     C.int(response.Status.Code),
				errorMsg: C.CString(response.Status.ErrorMsg),
			},
		}
	}

	if response.Status.ErrorMsg != "" {
		return C.MainAppInfo{
			status: C.Status{
				code:     C.int(response.Status.Code),
				errorMsg: C.CString(response.Status.ErrorMsg),
			},
		}
	}

	// 成功僅回傳 code
	status := C.Status{
		code: C.int(response.Status.Code),
	}

	replyContent := C.MainAppReplyContent{
		fileSize:   C.int64_t(response.MainAppInfo.ReplyContent.FileSize),
		fileName:   C.CString(response.MainAppInfo.ReplyContent.FileName),
		fileType:   C.CString(response.MainAppInfo.ReplyContent.FileType),
		appContent: C.CString(string(response.MainAppInfo.ReplyContent.AppContent)),
		version:    C.CString(response.MainAppInfo.ReplyContent.Version),
		blVersion:  C.CString(response.MainAppInfo.ReplyContent.BLVersion),
	}

	return C.MainAppInfo{
		askReplyId:      C.CString(response.MainAppInfo.AskReplyId),
		timeStampClient: C.CString(response.MainAppInfo.TimeStampClient),
		timeStampServer: C.CString(response.MainAppInfo.TimeStampServer),
		serverVersion:   C.CString(response.MainAppInfo.ServerVersion),
		msgValidateCode: C.CString(response.MainAppInfo.MsgValidateCode),
		replyContent:    replyContent,
		status:          status,
	}
}

func main() {}
