package hostinfo

type DataOverview struct {
	// 系统信息
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Kernel    string `json:"kernel"`
	Arch      string `json:"arch"`
	StartTime uint64 `json:"startTime"`
	RunTime   uint64 `json:"runTime"`

	Disks []Disk `json:"disks"`
}

type Memory struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
	Cache int64 `json:"cache"`

	SwapTotal int64 `json:"swapTotal"`
	SwapFree  int64 `json:"swapFree"`
	SwapUsed  int64 `json:"swapUsed"`
	SwapCache int64 `json:"swapCache"`
}

type Cpu struct {
	Index string `json:"index"`
	Total int64  `json:"total"`
	Idle  int64  `json:"idle"`
}

type Network struct {
	Name     string `json:"name"`
	Receive  int64  `json:"receive"`
	Transmit int64  `json:"transmit"`
}

type Disk struct {
	Total      int64  `json:"total"`
	Used       int64  `json:"used"`
	Free       int64  `json:"free"`
	FileSystem string `json:"fileSystem"`
	MountPoint string `json:"mountPoint"`
}

type TrendsCount struct {
	Day         string `json:"day" gorm:"column:day"`
	CreateCount int64  `json:"createCount" gorm:"column:count"`
}
