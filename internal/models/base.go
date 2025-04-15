package models

// 站點資訊
type Station struct {
	StationType string `json:"stationType"`
	StationName string `json:"stationName"`
	StationId   string `json:"stationId"`
	OperatorId  string `json:"operatorId"`
}

type Status struct {
	Code     int    `json:"code"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

type AskContent struct {
	ProductSeries     string `json:"productSeries"`
	ApplicableProject string `json:"applicableProject"`
	CustomizeId       string `json:"customizeId"`
}

type Request struct {
	Command         string     `json:"command"`
	TimeStampClient string     `json:"timeStampClient"`
	AskId           string     `json:"askId"`
	AskContent      AskContent `json:"askContent"`
	IsGetBinFile    bool       `json:"isGetBinFile"`
	Station         Station    `json:"station"`
	MsgValidateCode string     `json:"msgValidateCode"`
}

type Response struct {
	FileInfo    FileResponse    `json:"fileInfo"`
	MainAppInfo MainAppResponse `json:"MainAppInfo"`
	Status      Status          `json:"status"`
}

type ProcessRequest struct {
	TimeStamp       string     `json:"timeStamp"`
	AskId           string     `json:"askId"`
	AskContent      AskContent `json:"askContent"`
	Station         Station    `json:"station"`
	MsgValidateCode string     `json:"msgValidateCode"`
	SourcePath      string     `json:"sourcePath"`
}

const (
	// Server 端錯誤 (1xxx)
	ErrServerJSONMarshal         = 1001 // JSON 序列化錯誤
	ErrServerJSONUnmarshal       = 1002 // JSON 反序列化錯誤
	ErrServerWriteFile           = 1003 // 檔案寫入錯誤
	ErrServerReadFile            = 1004 // 檔案讀取錯誤
	ErrServerCreateDir           = 1005 // 目錄創建錯誤
	ErrServerFileSearch          = 1006 // 檔案搜尋錯誤
	ErrServerTimeoutResponseFile = 1007 // 等待回傳檔案超時
	ErrServerTimeoutStart        = 1008 // 服務器啟動超時
	ErrServerTimeoutTLSLoad      = 1009 // 加載 TLS 證書超時
	ErrServerConnectionAccept    = 1010 // 接受連接錯誤

	// Client 端錯誤 (2xxx)
	ErrClientNotConnected          = 2001 // 未連接
	ErrClientInvalidStation        = 2002 // 無效的站點資訊（從服務端返回）
	ErrClientInvalidAskId          = 2003 // 無效的 askId（從服務端返回）
	ErrClientConnectionInitFailed  = 2004 // 初始化連接失敗
	ErrClientRequestSendFailed     = 2005 // 發送請求失敗
	ErrClientResponseDecodeFailed  = 2006 // 回應解碼失敗
	ErrClientConnectionCloseFailed = 2007 // 關閉連接失敗
	ErrClientReadCACertFailed      = 2008 // 讀取 CA 證書失敗
	ErrClientParseCACertFailed     = 2009 // 解析 CA 證書失敗
	ErrClientTLSConnectionFailed   = 2010 // 建立 TLS 連接失敗
)
