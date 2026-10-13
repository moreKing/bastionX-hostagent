package hostinfo

import (
	"fmt"
	"host-agent/logger"
	"strconv"
	"sort"

	"github.com/shirou/gopsutil/v4/cpu"
)

// GetCPU 通过 gopsutil 获取每颗逻辑 CPU 的总时钟与空闲时钟（单位：秒，取整）。
func GetCPU() (*[]Cpu, error) {
	times, err := cpu.Times(true)
	if err != nil {
		fmt.Println("GetCPUInfo", "获取 CPU 时间信息失败", err)
		return nil, err
	}

	var cpuInfos []Cpu
	for i, t := range times {
		total := t.User + t.Nice + t.System + t.Idle + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Guest + t.GuestNice
		idle := t.Idle + t.Iowait

		cpuInfos = append(cpuInfos, Cpu{
			Index: fmt.Sprintf("cpu%d", i),
			Total: int64(total),
			Idle:  int64(idle),
		})
	}

	return &cpuInfos, nil
}

type CpuInfoTrends struct {
	Total CircularQueue `json:"total"`
	Idle  CircularQueue `json:"idle"`
}

var lastTenMinuteCpuInfo map[string]*CpuInfoTrends = make(map[string]*CpuInfoTrends)

func UpdateCpuInfoTrends() {
	cpuInfos, err := GetCPU()
	if err != nil {
		logger.Error("UpdateCpuInfoTrends", err)
		return
	}

	for _, v := range *cpuInfos {
		if _, exists := lastTenMinuteCpuInfo[v.Index]; !exists {
			lastTenMinuteCpuInfo[v.Index] = &CpuInfoTrends{
				Total: *NewCircularQueue(10),
				Idle:  *NewCircularQueue(10),
			}
		}

		// 更新cpu使用率趋势数据
		lastTenMinuteCpuInfo[v.Index].Total.Add(uint64(v.Total), true)
		lastTenMinuteCpuInfo[v.Index].Idle.Add(uint64(v.Idle), true)

		// 打印cpu使用率
		tmpTotal := float64(lastTenMinuteCpuInfo[v.Index].Total.LastValue())
		tmpIdle := float64(lastTenMinuteCpuInfo[v.Index].Idle.LastValue())

		if tmpTotal > 0 &&
			tmpIdle > 0 {

			logger.Debug("tmpIdle: ", tmpIdle, " tmpTotal: ", tmpTotal)
			logger.Info("Cpu 使用率 ", v.Index, fmt.Sprintf(" %.2f %%", (tmpTotal-tmpIdle)/tmpTotal*100))
		} else {
			logger.Warn("Cpu 使用率 ", v.Index, "计算失败，可能是首次计算或者数据异常", fmt.Sprintf("Total: %d, Idle: %d", v.Total, v.Idle))
		}

		logger.Info(v.Index, "Total:", v.Total, "Idle:", v.Idle)

		logger.Debug(lastTenMinuteCpuInfo[v.Index])
	}
}

type CpuInfoTrendsResponse struct {
	Index string   `json:"index"`
	Total []uint64 `json:"total"`
	Idle  []uint64 `json:"idle"`
}

// GetCpuUsageTrends 获取cpu使用率趋势数据
func GetCpuUsageTrends() *[]CpuInfoTrendsResponse {
	var response []CpuInfoTrendsResponse
	for name, v := range lastTenMinuteCpuInfo {
		resp := CpuInfoTrendsResponse{
			Index: name,
			Total: v.Total.GetAll(),
			Idle:  v.Idle.GetAll(),
		}
		response = append(response, resp)
	}

	// 根据cpu index排序
	sort.Slice(response, func(i, j int) bool {
		// 提取数字并转为 int 比较
		numI, _ := strconv.Atoi(response[i].Index[3:])
		numJ, _ := strconv.Atoi(response[j].Index[3:])
		return numI < numJ
	})

	return &response
}
