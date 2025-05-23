package models

type Station struct {
	StationType string `json:"stationType"`
	StationName string `json:"stationName"`
	StationId   string `json:"stationId"`
	OperatorId  string `json:"operatorId"`
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

type AskContent struct {
	ProductSeries     string `json:"productSeries"`
	ApplicableProject string `json:"applicableProject"`
	CustomizeId       string `json:"customizeId"`
}

type Response struct {
	FileInfo           FileResponse           `json:"fileInfo"`
	MainAppInfo        MainAppResponse        `json:"mainAppInfo"`
	ParameterTableInfo ParameterTableResponse `json:"parameterTableInfo"`
	Error              string                 `json:"error,omitempty"`
}
