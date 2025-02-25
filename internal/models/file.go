package models

type FileRequest struct {
	AskId             string `json:"askId"`
	ProductSeries     string `json:"productSeries"`
	ApplicableProject string `json:"applicableProject"`
	CustomizeId       string `json:"customized"`
}

type FileResponse struct {
	FileName    string `json:"fileName"`
	FilePath    string `json:"filePath"`
	FileSize    int64  `json:"fileSize"`
	FileContent []byte `json:"fileContent"`
}
