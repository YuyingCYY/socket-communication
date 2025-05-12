package models

// 站點資訊
type Station struct {
	StationType string `json:"stationType"`
	StationName string `json:"stationName"`
	StationId   string `json:"stationId"`
	OperatorId  string `json:"operatorId"`
}

type Status struct {
	ErrCode int    `json:"errCode"`
	ErrDesc string `json:"errDesc,omitempty"`
}

type AskContent struct {
	ProductSeries     string `json:"productSeries"`
	ApplicableProject string `json:"applicableProject"`
	CustomizeId       string `json:"customizeId"`
}

type Command string

const (
	Connect    Command = "connect"
	Disconnect Command = "disconnect"
	Download   Command = "download"
	Upload     Command = "upload"
)

type Request struct {
	Command           Command `json:"command"`
	TimeStampClient   string  `json:"timeStampClient"`
	AskId             string  `json:"askId"`
	ProductSeries     string  `json:"productSeries"`
	ApplicableProject string  `json:"applicableProject"`
	CustomizeId       string  `json:"customizeId"`
	StationType       string  `json:"stationType"`
	StationName       string  `json:"stationName"`
	StationId         string  `json:"stationId"`
	OperatorId        string  `json:"operatorId"`
	MsgValidateCode   string  `json:"msgValidateCode"`
	SerialNumber      uint16  `json:"serialNumber"`
}

type Response struct {
	MainAppInfo        MainAppResponse        `json:"MainAppInfo"`
	ParameterTableInfo ParameterTableResponse `json:"parameterTableInfo"`
	ProductInfo        ProductInfo            `json:"productInfo"`
	Status             Status                 `json:"status"`
}

type ProcessRequest struct {
	Command           Command `json:"command"`
	TimeStamp         string  `json:"timeStamp"`
	AskId             string  `json:"askId"`
	ProductSeries     string  `json:"productSeries"`
	ApplicableProject string  `json:"applicableProject"`
	CustomizeId       string  `json:"customizeId"`
	StationType       string  `json:"stationType"`
	StationName       string  `json:"stationName"`
	StationId         string  `json:"stationId"`
	OperatorId        string  `json:"operatorId"`
	MsgValidateCode   string  `json:"msgValidateCode"`
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
	ErrClientCertLoadFailed        = 2011 // 加載 CA 證書失敗
)
