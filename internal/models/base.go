package models

type Request struct {
	Command           string `json:"command"`
	TimeStampClient   string `json:"timeStampClient"`
	AskId             string `json:"askId"`
	ProductSeries     string `json:"productSeries"`
	ApplicableProject string `json:"applicableProject"`
	CustomizeId       string `json:"customizeId"`
	IsGetBinFile      bool   `json:"isGetBinFile"`
}

type Response struct {
	FileInfo    FileResponse    `json:"fileInfo"`
	MainAppInfo MainAppResponse `json:"MainAppInfo"`
	Error       string          `json:"error,omitempty"`
}

type ReplyContent struct {
	FileSize   int64  `json:"fileSize"`
	AppContent string `json:"appContent"`
	Version    string `json:"version"`
	BLVersion  string `json:"blVersion"`
}
