package models

type ShieldedZone struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type ParameterTableReplyContent struct {
	FileSize          int64          `json:"fileSize"`
	ParameterContent  []byte         `json:"parameterContent"`
	Version           string         `json:"version"`
	BLVersion         string         `json:"blVersion"`
	ShieldedZoneCount int            `json:"shieldedZoneCount"`
	ShieldedZone      []ShieldedZone `json:"shieldedZone"`
}

type ParameterTableResponse struct {
	AskReplyId      string                     `json:"askReplyId"`
	TimeStampClient string                     `json:"timeStampClient"`
	TimeStampServer string                     `json:"timeStampServer"`
	ServerVersion   string                     `json:"serverVersion"`
	MsgValidateCode string                     `json:"msgValidateCode"`
	ReplyContent    ParameterTableReplyContent `json:"replyContent"`
}
