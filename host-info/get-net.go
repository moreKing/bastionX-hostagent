package hostinfo

import (
	"host-agent/logger"
	"os"
	"path/filepath"
	"sort"

	"github.com/shirou/gopsutil/v4/net"
)

// GetNetInfo 获取网络信息
func GetNetInfo() (*[]Network, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Error("GetNetInfo", err)
		return nil, err
	}

	physicalNames := make(map[string]struct{})
	for _, iface := range ifaces {
		if iface.Name == "" || iface.Name == "lo" {
			continue
		}

		// 判断是否物理网卡
		devicePath := filepath.Join("/sys/class/net", iface.Name, "device")
		if _, statErr := os.Stat(devicePath); statErr != nil {
			logger.Debug(iface.Name, " 非物理网卡，跳过统计")
			continue
		}

		physicalNames[iface.Name] = struct{}{}
	}

	if len(physicalNames) == 0 {
		return &[]Network{}, nil
	}

	ioCounters, err := net.IOCounters(true)
	if err != nil {
		logger.Error("GetNetInfo", err)
		return nil, err
	}

	var netList []Network
	for _, counter := range ioCounters {
		if _, ok := physicalNames[counter.Name]; !ok {
			continue
		}

		logger.Info("GetNetInfo", counter.Name, "Receive:", counter.BytesRecv, "Transmit:", counter.BytesSent)
		netList = append(netList, Network{
			Name:     counter.Name,
			Receive:  int64(counter.BytesRecv),
			Transmit: int64(counter.BytesSent),
		})
	}

	// 按名称排序，保持输出稳定
	sort.Slice(netList, func(i, j int) bool { return netList[i].Name < netList[j].Name })

	return &netList, nil
}

type NetInfoTrends struct {
	Receive  CircularQueue `json:"receive"`
	Transmit CircularQueue `json:"transmit"`
}

type NetInfoTrend struct {
	Name     string `json:"name"`
	Receive  []uint `json:"receive"`
	Transmit []uint `json:"transmit"`
}

var lastTenMinuteNetInfo map[string]*NetInfoTrends = make(map[string]*NetInfoTrends)

// 每1分钟获取一次网络信息
func UpdateNetInfoTrends() {
	netInfo, err := GetNetInfo()
	if err != nil {
		logger.Error("UpdateNetInfoTrends", err)
		return
	}

	for _, v := range *netInfo {
		if _, exists := lastTenMinuteNetInfo[v.Name]; !exists {
			lastTenMinuteNetInfo[v.Name] = &NetInfoTrends{
				Receive:  CircularQueue{data: make([]uint64, 10)},
				Transmit: CircularQueue{data: make([]uint64, 10)},
			}
		}

		lastTenMinuteNetInfo[v.Name].Receive.Add(uint64(v.Receive), true)
		lastTenMinuteNetInfo[v.Name].Transmit.Add(uint64(v.Transmit), true)

		logger.Info("UpdateNetInfoTrends", v.Name, "Receive:", v.Receive, "Transmit:", v.Transmit)
		logger.Debug(lastTenMinuteNetInfo[v.Name])
	}
}

type NetInfoTrendsResponse struct {
	Name          string `json:"name"`
	TotalReceive  uint64 `json:"totalReceive"`
	TotalTransmit uint64 `json:"totalTransmit"`

	ReceiveSpeed  []uint64 `json:"receiveSpeed"`
	TransmitSpeed []uint64 `json:"transmitSpeed"`
}

// GetNetInfoTrends 获取网络信息趋势数据
func GetNetInfoTrends() *[]NetInfoTrendsResponse {
	var response []NetInfoTrendsResponse
	for name, v := range lastTenMinuteNetInfo {
		resp := NetInfoTrendsResponse{
			Name:          name,
			TotalReceive:  v.Receive.lastValue,
			TotalTransmit: v.Transmit.lastValue,

			ReceiveSpeed:  v.Receive.GetAll(),
			TransmitSpeed: v.Transmit.GetAll(),
		}
		response = append(response, resp)
	}
	return &response
}
