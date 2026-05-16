package hostinfo

import (
	"host-agent/logger"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var reNetInfo = regexp.MustCompile(`(?m)^([^:\s]+):\s+(\d+)\s+(\d+)`)

// GetNetInfo 获取网络信息
func GetNetInfo() (*[]Network, error) {

	// 获取网络信息
	//	output, err := session.Output(`awk 'NR>2 {print $1, $2, $10}' /proc/net/dev`)
	output, err := exec.Command("awk", "NR>2 {print $1, $2, $10}", "/proc/net/dev").Output()
	if err != nil {
		logger.Error("GetNetInfo", err)
		return nil, err
	}

	var netList []Network
	for _, line := range reNetInfo.FindAllStringSubmatch(string(output), -1) {
		logger.Info("GetNetInfo", line)
		rx, _ := strconv.ParseInt(strings.TrimSpace(line[2]), 10, 64)
		tx, _ := strconv.ParseInt(strings.TrimSpace(line[3]), 10, 64)

		netList = append(netList, Network{
			Name:     line[1],
			Receive:  rx,
			Transmit: tx,
		})
	}

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
