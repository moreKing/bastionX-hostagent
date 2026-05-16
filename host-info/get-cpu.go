package hostinfo

import (
	"fmt"
	"host-agent/logger"
	"os/exec"
	"strconv"
	"strings"
)

func GetCPU() (*[]Cpu, error) {

	output, err := exec.Command("cat", "/proc/stat").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		fmt.Println("GetCPUInfo", "解析CPU信息失败", lines)
		return nil, fmt.Errorf("解析CPU信息失败")
	}

	var cpuInfos []Cpu

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "cpu") && !strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 10 {
				fmt.Println("GetCpuUsageRate", "解析cpu信息失败", line)
				continue
			}
			// 字符串转数字
			user, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu user信息失败", fields[1], err)
				continue
			}
			nice, err := strconv.ParseInt(fields[2], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu nice信息失败", fields[2], err)
				continue
			}
			system, err := strconv.ParseInt(fields[3], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu system信息失败", fields[3], err)
				continue
			}
			idle, err := strconv.ParseInt(fields[4], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu idle信息失败", fields[4], err)
				continue
			}
			iowait, err := strconv.ParseInt(fields[5], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu iowait信息失败", fields[5], err)
				continue
			}
			irq, err := strconv.ParseInt(fields[6], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu irq信息失败", fields[6], err)
				continue
			}
			softirq, err := strconv.ParseInt(fields[7], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu softirq信息失败", fields[7], err)
				continue
			}
			steal, err := strconv.ParseInt(fields[8], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu steal信息失败", fields[8], err)
				continue
			}
			guest, err := strconv.ParseInt(fields[9], 10, 64)
			if err != nil {
				fmt.Println("GetCpuUsageRate", "解析cpu guest信息失败", fields[9], err)
				continue
			}
			// 计算cpu使用率
			cpuInfos = append(cpuInfos, Cpu{
				Index: fields[0],
				Total: user + nice + system + idle + iowait + irq + softirq + steal + guest,
				Idle:  idle + iowait,
			})
		}
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
				Total: CircularQueue{data: make([]uint64, 10)},
				Idle:  CircularQueue{data: make([]uint64, 10)},
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
	return &response
}
