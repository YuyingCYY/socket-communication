package main

/*
#include <stdlib.h>

// 定義結構體（與 Go 中的定義一致）
typedef struct {
	char* fileName;
	char* filePath;
	long fileSize;
	char* fileContent;
	size_t fileContentSize;
	char* errorMsg;
} BinFileInfo;

typedef struct {
	long fileSize;
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
	"unsafe"

	"github.com/socket-communication/internal/client"
	"github.com/socket-communication/internal/models"
)

//export InitConnection
func InitConnection() bool {
	err := client.InitConnection()
	return err == nil
}

//export CloseConnection
func CloseConnection() *C.char {
	request := models.Request{
		Command: "disconnect",
	}
	_, err := client.SendRequest(request)
	if err != nil {
		response := models.Response{
			Error: err.Error(),
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	return C.CString("disconnect")
	// client.CloseConnection()
}

//export GetBinFileInfo
func GetBinFileInfo(askId, productSeries, applicableProject, customizeId *C.char) C.BinFileInfo {
	request := models.Request{
		AskId:             C.GoString(askId),
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
		IsGetBinFile:      true,
	}

	decoder, err := client.SendRequest(request)
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

	return C.BinFileInfo{
		fileName:        C.CString(response.FileInfo.FileName),
		filePath:        C.CString(response.FileInfo.FilePath),
		fileSize:        C.long(response.FileInfo.FileSize),
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
	if info.error != nil {
		C.free(unsafe.Pointer(info.error))
	}
}

//export GetMainAppInfo
func GetMainAppInfo(askId, productSeries, applicableProject, customizeId *C.char) C.MainAppInfo {
	request := models.Request{
		AskId:             C.GoString(askId),
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
		IsGetBinFile:      false,
	}

	decoder, err := client.SendRequest(request)
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

	replyContent := C.ReplyContent{
		fileSize:   C.long(response.MainAppInfo.ReplyContent.FileSize),
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
