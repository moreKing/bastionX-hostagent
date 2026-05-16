package hostinfo

import (
	"host-agent/logger"
	"time"
)

type SystemAllTrendsResponse struct {
	CpuTrends *[]CpuInfoTrendsResponse `json:"cpuTrends"`
	MemTrends *MemoryTrendsResponse    `json:"memTrends"`
	NetTrends *[]NetInfoTrendsResponse `json:"netTrends"`

	Disks  *[]Disk     `json:"disks"`
	System *SystemInfo `json:"system"`

	Timestamp int64 `json:"timestamp"`
}

func GetSystemAllTrends() (*SystemAllTrendsResponse, error) {

	disks, err := GetDisk()
	if err != nil {
		logger.Error("获取磁盘信息失败： ", err)
		return nil, err
	}

	sysInfo, err := GetSystemInfo()
	if err != nil {
		logger.Error("获取系统信息失败： ", err)
		return nil, err
	}

	return &SystemAllTrendsResponse{
		CpuTrends: GetCpuUsageTrends(),
		MemTrends: GetMemoryTrends(),
		NetTrends: GetNetInfoTrends(),

		Disks:  disks,
		System: sysInfo,

		Timestamp: time.Now().UnixMilli(),
	}, nil

}
