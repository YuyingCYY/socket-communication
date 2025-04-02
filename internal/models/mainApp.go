package models

type MainAppReplyContent struct {
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	FileType   string `json:"fileType"`
	AppContent []byte `json:"appContent"`
	Version    string `json:"version"`
	BLVersion  string `json:"blVersion"`
}

type MainAppResponse struct {
	AskReplyId      string              `json:"askReplyId"`
	TimeStampClient string              `json:"timeStampClient"`
	TimeStampServer string              `json:"timeStampServer"`
	ServerVersion   string              `json:"serverVersion"`
	ReplyStatus     string              `json:"replyStatus"`
	MsgValidateCode string              `json:"msgValidateCode"`
	ReplyContent    MainAppReplyContent `json:"replyContent"`
}
