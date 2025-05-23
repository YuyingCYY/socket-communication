package models

type MainAppResponse struct {
	AskReplyId      string       `json:"askReplyId"`
	TimeStampClient string       `json:"timeStampClient"`
	TimeStampServer string       `json:"timeStampServer"`
	ServerVersion   string       `json:"serverVersion"`
	ReplyStatus     string       `json:"replyStatus"`
	MsgValidateCode string       `json:"msgValidateCode"`
	ReplyContent    ReplyContent `json:"replyContent"`
}
