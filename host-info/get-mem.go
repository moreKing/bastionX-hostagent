package hostinfo

import (
	"fmt"
	"host-agent/logger"

	"github.com/shirou/gopsutil/v4/mem"
)

func GetMemory() (*Memory, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("GetMemInfo", "获取虚拟内存信息失败", err)
		return nil, err
	}
	swap, err := mem.SwapMemory()
	if err != nil {
		fmt.Println("GetMemInfo", "获取交换内存信息失败", err)
		return nil, err
	}

	total := int64(vm.Total)
	free := int64(vm.Free)
	cache := int64(vm.Buffers + vm.Cached)
	used := int64(vm.Total) - free - cache
	if used < 0 {
		used = int64(vm.Used)
	}

	memInfo := Memory{
		Total:     total,
		Used:      used,
		Free:      free,
		Cache:     cache,
		SwapTotal: int64(swap.Total),
		SwapFree:  int64(swap.Free),
		SwapUsed:  int64(swap.Used),
		SwapCache: 0,
	}

	return &memInfo, nil
}

type PhysicalMemory struct {
	Total int64         `json:"total"`
	Used  CircularQueue `json:"used"`
	Cache CircularQueue `json:"cache"`
}

var lastTenMinuteMemInfo = PhysicalMemory{
	Used:  CircularQueue{data: make([]uint64, 10)},
	Cache: CircularQueue{data: make([]uint64, 10)},
}

// 每1分钟获取一次内存信息
func UpdateMemoryTrends() {
	memInfo, err := GetMemory()
	if err != nil {
		logger.Error("获取内存数据失败： ", err)
		return
	}

	if memInfo.Used > 0 && memInfo.Cache > 0 {
		logger.Info("内存使用率 ", fmt.Sprintf(" %.2f %%", float64(memInfo.Used)/float64(memInfo.Total)*100))
	} else {
		logger.Error("内存使用率计算失败", fmt.Sprintf("Used: %d, Cache: %d", memInfo.Used, memInfo.Cache))
	}

	logger.Info("UpdateMemoryTrends", "Total:", memInfo.Total, "Used:", memInfo.Used, "Cache:", memInfo.Cache)

	lastTenMinuteMemInfo.Total = memInfo.Total
	lastTenMinuteMemInfo.Used.Add(uint64(memInfo.Used))
	lastTenMinuteMemInfo.Cache.Add(uint64(memInfo.Cache))

	logger.Debug(lastTenMinuteMemInfo)
}

type MemoryTrendsResponse struct {
	Total int64    `json:"total"`
	Used  []uint64 `json:"used"`
	Cache []uint64 `json:"cache"`
}

func GetMemoryTrends() *MemoryTrendsResponse {
	return &MemoryTrendsResponse{
		Total: lastTenMinuteMemInfo.Total,
		Used:  lastTenMinuteMemInfo.Used.GetAll(),
		Cache: lastTenMinuteMemInfo.Cache.GetAll(),
	}
}
