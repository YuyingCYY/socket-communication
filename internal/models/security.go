package models

type ProductInfo struct {
	ProductName     string   `json:"productName"`
	ServeNo         string   `json:"serveNo"`
	ManufactWeek    uint8    `json:"manufactWeek"`
	ManufactYear    uint8    `json:"manufactYear"`
	Echelon         uint16   `json:"echelon"`
	SerialNumber    uint16   `json:"serialNumber"`
	CheckCode       uint16   `json:"checkCode"`
	WarrantyWeek    uint8    `json:"warrantyWeek"`
	WarrantyYear    uint8    `json:"warrantyYear"`
	SysEnable       uint8    `json:"sysEnable"`
	SysMatchState   uint8    `json:"sysMatchState"`
	SysMatchNum     uint8    `json:"sysMatchNum"`
	SysMatchInfo    []uint64 `json:"sysMatchInfo"`
	WarrantyState   uint8    `json:"warrantyState"`
	WarrantyTimeCnt uint32   `json:"warrantyTimecnt"`
}
