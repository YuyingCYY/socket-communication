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

typedef struct {
	int start;
	int end;
} ShieldedZone;

typedef struct {
	char* fileName;
	int64_t fileSize;
	char* fileType;
	char* appContent;
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
	Status status;
} ParameterTableInfo;

*/
import "C"
import (
	"strconv"
	"time"

	"github.com/socket-communication/internal/client"
	"github.com/socket-communication/internal/models"
)

//export InitConnection
func InitConnection(stationType, stationName, stationId, operatorId *C.char) C.Status {
	errCode, err := client.GetClient().InitConnection()
	if err != nil {
		return C.Status{
			code:     C.int(errCode), // 初始化連接失敗通常與 TLS 相關
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

//export GetParameterTableInfo
func GetParameterTableInfo(productSeries, applicableProject, customizeId *C.char) C.ParameterTableInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.ParameterTableInfo{
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
		AskId:           "ParameterTable",
		AskContent:      askContent,
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.ParameterTableInfo{
			status: C.Status{
				code:     C.int(models.ErrClientRequestSendFailed),
				errorMsg: C.CString(err.Error()),
			},
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.ParameterTableInfo{
			status: C.Status{
				code:     C.int(models.ErrClientResponseDecodeFailed),
				errorMsg: C.CString(err.Error()),
			},
		}
	}

	// 代碼不為`0`皆為錯誤碼
	if response.Status.Code != 0 {
		return C.ParameterTableInfo{
			status: C.Status{
				code:     C.int(response.Status.Code),
				errorMsg: C.CString(response.Status.ErrorMsg),
			},
		}
	}

	if response.Status.ErrorMsg != "" {
		return C.ParameterTableInfo{
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

	shieldedZone := response.ParameterTableInfo.ReplyContent.ShieldedZone
	var cShieldedZones [10]C.ShieldedZone

	for i := range response.ParameterTableInfo.ReplyContent.ShieldedZoneCount {
		cShieldedZones[i].start = C.int(shieldedZone[i].Start)
		cShieldedZones[i].end = C.int(shieldedZone[i].End)
	}

	replyContent := C.ParameterTableReplyContent{
		fileSize:          C.int64_t(response.ParameterTableInfo.ReplyContent.FileSize),
		fileName:          C.CString(response.ParameterTableInfo.ReplyContent.FileName),
		fileType:          C.CString(response.ParameterTableInfo.ReplyContent.FileType),
		appContent:        C.CString(string(response.ParameterTableInfo.ReplyContent.AppContent)),
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
		status:          status,
	}
}

//export FaultLog
func FaultLog(errCode C.int) {

}

func main() {}
