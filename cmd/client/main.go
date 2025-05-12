package main

/*
#include <stdlib.h>
#include <stdint.h>

// 定義結構體（與 Go 中的定義一致）
typedef struct {
	int errCode;
	char* errDesc;
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
	Status status;
} ParameterTableInfo;

typedef struct {
	char* productName;
	char* serveNo;
	uint8_t maunfactWeek;
	uint8_t maunfactYear;
	uint16_t echelon;
	uint16_t serialNumber;
	uint16_t checkCode;
	uint8_t warrantyWeek;
	uint8_t warrantyYear;
	uint8_t sysEnable;
	uint8_t sysMatchState;
	uint8_t sysMatchNum;
	uint64_t sysMatchInfo[12];
	uint8_t warrantyState;
	uint8_t warrantyTimecnt;
	Status status;
} ProductInfo;
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
	errCode, err := client.GetClient().InitConnection()
	if err != nil {
		return C.Status{
			errCode: C.int(errCode), // 初始化連接失敗通常與 TLS 相關
			errDesc: C.CString(err.Error()),
		}
	}

	request := models.Request{
		Command:     models.Connect,
		StationType: C.GoString(stationType),
		StationName: C.GoString(stationName),
		StationId:   C.GoString(stationId),
		OperatorId:  C.GoString(operatorId),
	}
	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.Status{
			errCode: models.ErrClientRequestSendFailed,
			errDesc: C.CString(err.Error()),
		}
	}

	var response models.Status
	if err := decoder.Decode(&response); err != nil {
		return C.Status{
			errCode: models.ErrClientResponseDecodeFailed,
			errDesc: C.CString(err.Error()),
		}
	}

	return C.Status{
		errCode: 0,
	}
}

//export CloseConnection
func CloseConnection() C.Status {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.Status{
			errCode: models.ErrClientNotConnected,
			errDesc: C.CString("Not connected"),
		}
	}

	request := models.Request{
		Command: models.Disconnect,
	}
	_, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.Status{
			errCode: models.ErrClientRequestSendFailed,
			errDesc: C.CString(err.Error()),
		}
	}

	if err := client.GetClient().CloseConnection(); err != nil {
		return C.Status{
			errCode: models.ErrClientConnectionCloseFailed,
			errDesc: C.CString(err.Error()),
		}
	}

	return C.Status{
		errCode: 0,
	}
}

//export GetMainAppInfo
func GetMainAppInfo(productSeries, applicableProject, customizeId *C.char) C.MainAppInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.MainAppInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientNotConnected),
				errDesc: C.CString("Not connected"),
			},
		}
	}

	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)

	request := models.Request{
		Command:           models.Download,
		TimeStampClient:   timeUnix,
		AskId:             "MainApp",
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.MainAppInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientRequestSendFailed),
				errDesc: C.CString(err.Error()),
			},
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.MainAppInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientResponseDecodeFailed),
				errDesc: C.CString(err.Error()),
			},
		}
	}

	// 代碼不為`0`皆為錯誤碼
	if response.Status.ErrCode != 0 {
		return C.MainAppInfo{
			status: C.Status{
				errCode: C.int(response.Status.ErrCode),
				errDesc: C.CString(response.Status.ErrDesc),
			},
		}
	}

	if response.Status.ErrDesc != "" {
		return C.MainAppInfo{
			status: C.Status{
				errCode: C.int(response.Status.ErrCode),
				errDesc: C.CString(response.Status.ErrDesc),
			},
		}
	}

	// 成功僅回傳 code
	status := C.Status{
		errCode: C.int(response.Status.ErrCode),
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

//export FreeMainAppInfo
func FreeMainAppInfo(info *C.MainAppInfo) {
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
	C.free(unsafe.Pointer(info.replyContent.appContent))
	C.free(unsafe.Pointer(info.replyContent.version))
	C.free(unsafe.Pointer(info.replyContent.blVersion))
}

//export GetParameterTableInfo
func GetParameterTableInfo(productSeries, applicableProject, customizeId *C.char) C.ParameterTableInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.ParameterTableInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientNotConnected),
				errDesc: C.CString("Not connected"),
			},
		}
	}

	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)

	request := models.Request{
		Command:           models.Download,
		TimeStampClient:   timeUnix,
		AskId:             "ParameterTable",
		ProductSeries:     C.GoString(productSeries),
		ApplicableProject: C.GoString(applicableProject),
		CustomizeId:       C.GoString(customizeId),
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.ParameterTableInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientRequestSendFailed),
				errDesc: C.CString(err.Error()),
			},
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.ParameterTableInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientResponseDecodeFailed),
				errDesc: C.CString(err.Error()),
			},
		}
	}

	// 代碼不為`0`皆為錯誤碼
	if response.Status.ErrCode != 0 {
		return C.ParameterTableInfo{
			status: C.Status{
				errCode: C.int(response.Status.ErrCode),
				errDesc: C.CString(response.Status.ErrDesc),
			},
		}
	}

	if response.Status.ErrDesc != "" {
		return C.ParameterTableInfo{
			status: C.Status{
				errCode: C.int(response.Status.ErrCode),
				errDesc: C.CString(response.Status.ErrDesc),
			},
		}
	}

	// 成功僅回傳 code
	status := C.Status{
		errCode: C.int(response.Status.ErrCode),
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
		status:          status,
	}
}

//export GetProductInfo
func GetProductInfo(serialNumber *C.uint16_t) C.ProductInfo {
	// 檢查是否已連線
	if !client.GetClient().IsConnected() {
		return C.ProductInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientNotConnected),
				errDesc: C.CString("Not connected"),
			},
		}
	}

	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)

	request := models.Request{
		Command:         models.Download,
		TimeStampClient: timeUnix,
		AskId:           "SecurityZone",
		SerialNumber:    uint16(*serialNumber),
	}

	decoder, err := client.GetClient().SendRequest(request)
	if err != nil {
		return C.ProductInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientRequestSendFailed),
				errDesc: C.CString(err.Error()),
			},
		}
	}

	var response models.Response
	if err := decoder.Decode(&response); err != nil {
		return C.ProductInfo{
			status: C.Status{
				errCode: C.int(models.ErrClientResponseDecodeFailed),
				errDesc: C.CString(err.Error()),
			},
		}
	}

	// 代碼不為`0`皆為錯誤碼
	if response.Status.ErrCode != 0 {
		return C.ProductInfo{
			status: C.Status{
				errCode: C.int(response.Status.ErrCode),
				errDesc: C.CString(response.Status.ErrDesc),
			},
		}
	}

	if response.Status.ErrDesc != "" {
		return C.ProductInfo{
			status: C.Status{
				errCode: C.int(response.Status.ErrCode),
				errDesc: C.CString(response.Status.ErrDesc),
			},
		}
	}

	// 成功僅回傳 code
	status := C.Status{
		errCode: C.int(response.Status.ErrCode),
	}

	info := response.ProductInfo
	cProductInfo := C.ProductInfo{
		productName:     C.CString(info.ProductName),
		serveNo:         C.CString(info.ServeNo),
		maunfactWeek:    C.uint8_t(int(info.ManufactWeek)),
		maunfactYear:    C.uint8_t(info.ManufactYear),
		echelon:         C.uint16_t(info.Echelon),
		serialNumber:    C.uint16_t(info.SerialNumber),
		checkCode:       C.uint16_t(info.CheckCode),
		warrantyWeek:    C.uint8_t(info.WarrantyWeek),
		warrantyYear:    C.uint8_t(info.WarrantyYear),
		sysEnable:       C.uint8_t(info.SysEnable),
		sysMatchState:   C.uint8_t(info.SysMatchState),
		sysMatchNum:     C.uint8_t(info.SysMatchNum),
		warrantyState:   C.uint8_t(info.WarrantyState),
		warrantyTimecnt: C.uint8_t(info.WarrantyTimeCnt),
		status:          status,
	}
	for i := range info.SysMatchInfo {
		if i < len(cProductInfo.sysMatchInfo) {
			cProductInfo.sysMatchInfo[i] = C.uint64_t(info.SysMatchInfo[i])
		}
	}
	return cProductInfo
}

//export FaultLog
func FaultLog(errCode C.int) {

}

func main() {}
