package hostinfo

import (
	"fmt"
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
		fmt.Println("GetNetInfo", err)
		return nil, err
	}

	var netList []Network
	for _, line := range reNetInfo.FindAllStringSubmatch(string(output), -1) {
		fmt.Println("GetNetInfo", line)
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
