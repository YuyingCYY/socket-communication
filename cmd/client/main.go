package main

/*
#include <stdlib.h>
#include <stdint.h>

// 定義結構體（與 Go 中的定義一致）
typedef struct {
	int64_t fileSize;
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
	char* errorMsg;
} MainAppInfo;

typedef struct {
	int start;
	int end;
} ShieldedZone;

typedef struct {
	char* fileName;
	int64_t fileSize;
	char* fileType;
	char* parameterContent;
	char* version;
	char* blVersion;
	int shieldedZoneCount;
	ShieldedZone shieldedZone[10];
} ParameterTableReplyContent;

typedef struct {
	char* askReplyId;
	char* timeStampClient;
	char* timeStampServer;
	char* serverVersion;
	char* msgValidateCode;
	ParameterTableReplyContent replyContent;
	char* errorMsg;
} ParameterTableInfo;

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
		return C.CString(err.Error())
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
		jsonResponse, _ := json.Marshal(response)
		return C.CString("Error:" + string(jsonResponse))
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.CString("Error: " + err.Error())
	}

	if response.Error != "" {
		return C.CString("Error: " + response.Error)
	}

	return C.CString("Initialize successfully")
}

//export CloseConnection
func CloseConnection() *C.char {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		response := models.Response{
			Error: "Not connected",
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
			Error: err.Error(),
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	if err := client.GetClient().CloseConnection(); err != nil {
		response := models.Response{
			Error: err.Error(),
		}
		jsonResponse, _ := json.Marshal(response)
		return C.CString(string(jsonResponse))
	}

	return C.CString("disconnect")
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

	replyContent := C.MainAppReplyContent{
		fileSize:   C.int64_t(response.MainAppInfo.ReplyContent.FileSize),
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
	}
}

//export GetParameterTableInfo
func GetParameterTableInfo(productSeries, applicableProject, customizeId, fileName *C.char) C.ParameterTableInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.ParameterTableInfo{
			errorMsg: C.CString("Not connected"),
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
		AskId:           "ParameterTable",
		AskContent:      askContent,
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.ParameterTableInfo{
			errorMsg: C.CString(err.Error()),
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.ParameterTableInfo{
			errorMsg: C.CString(err.Error()),
		}
	}

	if response.Error != "" {
		return C.ParameterTableInfo{
			errorMsg: C.CString(response.Error),
		}
	}

	shieldedZone := response.ParameterTableInfo.ReplyContent.ShieldedZone
	var cShieldedZones [10]C.ShieldedZone

	for i := range response.ParameterTableInfo.ReplyContent.ShieldedZoneCount {
		cShieldedZones[i].start = C.int(shieldedZone[i].Start)
		cShieldedZones[i].end = C.int(shieldedZone[i].End)
	}

	replyContent := C.ParameterTableReplyContent{
		fileSize:          C.int64_t(response.ParameterTableInfo.ReplyContent.FileSize),
		parameterContent:  C.CString(string(response.ParameterTableInfo.ReplyContent.ParameterContent)),
		version:           C.CString(response.ParameterTableInfo.ReplyContent.Version),
		blVersion:         C.CString(response.ParameterTableInfo.ReplyContent.BLVersion),
		shieldedZoneCount: C.int(response.ParameterTableInfo.ReplyContent.ShieldedZoneCount),
		shieldedZone:      cShieldedZones,
	}

	return C.ParameterTableInfo{
		askReplyId:      C.CString(response.ParameterTableInfo.AskReplyId),
		timeStampClient: C.CString(response.ParameterTableInfo.TimeStampClient),
		timeStampServer: C.CString(response.ParameterTableInfo.TimeStampServer),
		serverVersion:   C.CString(response.ParameterTableInfo.ServerVersion),
		msgValidateCode: C.CString(response.ParameterTableInfo.MsgValidateCode),
		replyContent:    replyContent,
	}
}

//export FreeParameterTableInfo
func FreeParameterTableInfo(info *C.ParameterTableInfo) {
	if info == nil {
		return
	}

	C.free(unsafe.Pointer(info.askReplyId))
	C.free(unsafe.Pointer(info.timeStampClient))
	C.free(unsafe.Pointer(info.timeStampServer))
	C.free(unsafe.Pointer(info.serverVersion))
	C.free(unsafe.Pointer(info.msgValidateCode))
	C.free(unsafe.Pointer(info.replyContent.fileName))
	C.free(unsafe.Pointer(info.replyContent.fileType))
	C.free(unsafe.Pointer(info.replyContent.parameterContent))
	C.free(unsafe.Pointer(info.replyContent.version))
	C.free(unsafe.Pointer(info.replyContent.blVersion))
	C.free(unsafe.Pointer(info.errorMsg))
}

func main() {}
